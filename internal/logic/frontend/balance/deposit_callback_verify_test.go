package balance

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"

	v1 "jh_balance_service/api/frontend/balance/v1"
	"jh_balance_service/internal/model/entity"
)

// cbMockDriver 复用 GoFrame MySQL 方言，底层连接替换为 sqlmock。
type cbMockDriver struct {
	*mysql.Driver
}

var (
	cbOnce sync.Once
	cbDB   *sql.DB
	cbMock sqlmock.Sqlmock
)

func (d *cbMockDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	return &cbMockDriver{Driver: &mysql.Driver{Core: core}}, nil
}

func (d *cbMockDriver) Open(config *gdb.ConfigNode) (*sql.DB, error) { return cbDB, nil }

// TableFields 返回空表结构，避免 SHOW FULL COLUMNS 查询。
func (d *cbMockDriver) TableFields(ctx context.Context, table string, schema ...string) (map[string]*gdb.TableField, error) {
	return map[string]*gdb.TableField{}, nil
}

func setupCallbackTest(t *testing.T, verify bool) sqlmock.Sqlmock {
	t.Helper()
	cbOnce.Do(func() {
		var err error
		cbDB, cbMock, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		if err = gdb.Register("sqlmock_cb", &cbMockDriver{}); err != nil {
			t.Fatalf("register driver: %v", err)
		}
	})
	adapter, err := gcfg.NewAdapterContent(fmt.Sprintf(
		`{"database":{"default":{"type":"sqlmock_cb","link":"sqlmock_cb:u:p@tcp(127.0.0.1:3306)/test"}},"payment":{"mg":{"verifyCallback":%v}}}`, verify))
	if err != nil {
		t.Fatalf("config adapter: %v", err)
	}
	g.Cfg().SetAdapter(adapter)
	t.Cleanup(func() {
		if err := cbMock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}
	})
	return cbMock
}

// mgSign 按 MG 签名规则（过滤空值、key 升序、追加 &key=、MD5 大写）独立计算签名。
func mgSign(tradeStatus, orderSn string, amount float64, key string) string {
	s := fmt.Sprintf("actual_amount=%.2f&order_sn=%s&trade_status=%s&key=%s", amount, orderSn, tradeStatus, key)
	h := md5.Sum([]byte(s))
	return strings.ToUpper(hex.EncodeToString(h[:]))
}

func expectChannel(mock sqlmock.Sqlmock, code, md5Key string) {
	mock.ExpectQuery("FROM `balance_log_payment`").
		WillReturnRows(sqlmock.NewRows([]string{"site_id", "payment_id", "trade_no"}).AddRow(1, 7, "T001"))
	mock.ExpectQuery("FROM `site_payment`").
		WillReturnRows(sqlmock.NewRows([]string{"code", "md5_key"}).AddRow(code, md5Key))
}

func newReq(amount float64, sign string) *v1.DepositCallbackReq {
	return &v1.DepositCallbackReq{TradeStatus: "success", OrderSn: "T001", ActualAmount: amount, Sign: sign}
}

var order = &entity.BalanceLog{Money: 100}

func TestVerifyDepositCallback_ValidSign(t *testing.T) {
	mock := setupCallbackTest(t, true)
	expectChannel(mock, "MG", "secret")
	req := newReq(100, mgSign("success", "T001", 100, "secret"))
	if !New().verifyDepositCallback(context.Background(), req, order) {
		t.Fatal("expected valid callback to pass")
	}
}

func TestVerifyDepositCallback_BadSign(t *testing.T) {
	mock := setupCallbackTest(t, true)
	expectChannel(mock, "MG", "secret")
	req := newReq(100, mgSign("success", "T001", 100, "wrong-key"))
	if New().verifyDepositCallback(context.Background(), req, order) {
		t.Fatal("expected bad sign to be rejected")
	}
}

func TestVerifyDepositCallback_AmountAboveOrder(t *testing.T) {
	setupCallbackTest(t, true) // 金额校验在查库前，不应发出任何 SQL
	req := newReq(100.02, mgSign("success", "T001", 100.02, "secret"))
	if New().verifyDepositCallback(context.Background(), req, order) {
		t.Fatal("expected amount above order to be rejected")
	}
}

func TestVerifyDepositCallback_NonPositiveAmount(t *testing.T) {
	setupCallbackTest(t, true)
	for _, amt := range []float64{0, -1} {
		if New().verifyDepositCallback(context.Background(), newReq(amt, "x"), order) {
			t.Errorf("expected amount %v to be rejected", amt)
		}
	}
}

func TestVerifyDepositCallback_UnknownChannel(t *testing.T) {
	mock := setupCallbackTest(t, true)
	expectChannel(mock, "UNKNOWN_PAY", "secret")
	req := newReq(100, mgSign("success", "T001", 100, "secret"))
	if New().verifyDepositCallback(context.Background(), req, order) {
		t.Fatal("expected unknown channel to be rejected")
	}
}

func TestVerifyDepositCallback_ChannelNotFound(t *testing.T) {
	mock := setupCallbackTest(t, true)
	mock.ExpectQuery("FROM `balance_log_payment`").
		WillReturnRows(sqlmock.NewRows([]string{"site_id", "payment_id", "trade_no"}).AddRow(1, 7, "T001"))
	mock.ExpectQuery("FROM `site_payment`").WillReturnRows(sqlmock.NewRows([]string{"code", "md5_key"}))
	if New().verifyDepositCallback(context.Background(), newReq(100, "x"), order) {
		t.Fatal("expected missing channel config to be rejected")
	}
}

func TestVerifyDepositCallback_VerifyDisabledAmountOnly(t *testing.T) {
	setupCallbackTest(t, false) // 关闭验签时仅做金额比对，不查库
	if !New().verifyDepositCallback(context.Background(), newReq(100, ""), order) {
		t.Fatal("expected pass when verify disabled and amount ok")
	}
	if New().verifyDepositCallback(context.Background(), newReq(200, ""), order) {
		t.Fatal("expected amount check to still reject when verify disabled")
	}
}
