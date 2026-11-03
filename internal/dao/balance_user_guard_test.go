package dao

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// mockDriver 复用 GoFrame MySQL 方言，但底层连接替换为 sqlmock，测试无需真实数据库。
type mockDriver struct {
	*mysql.Driver
}

var (
	mockOnce sync.Once
	mockDB   *sql.DB
	mockSQL  sqlmock.Sqlmock
)

func (d *mockDriver) New(core *gdb.Core, node *gdb.ConfigNode) (gdb.DB, error) {
	return &mockDriver{Driver: &mysql.Driver{Core: core}}, nil
}

func (d *mockDriver) Open(config *gdb.ConfigNode) (*sql.DB, error) {
	return mockDB, nil
}

// TableFields 返回固定表结构，避免 SHOW FULL COLUMNS 查询。
func (d *mockDriver) TableFields(ctx context.Context, table string, schema ...string) (map[string]*gdb.TableField, error) {
	names := []string{"id", "site_id", "user_id", "username", "balance", "balance_frozen", "balance_debt",
		"points", "balance_status", "currency", "created_at", "updated_at"}
	m := make(map[string]*gdb.TableField, len(names))
	for i, n := range names {
		m[n] = &gdb.TableField{Index: i, Name: n, Type: "decimal(16,4)"}
	}
	return m, nil
}

func setupMock(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	mockOnce.Do(func() {
		// 使用测试配置，避免加载 manifest 中的真实数据库配置
		adapter, err := gcfg.NewAdapterContent(`{"database":{"default":{"type":"sqlmock_guard","link":"sqlmock_guard:u:p@tcp(127.0.0.1:3306)/test"}}}`)
		if err != nil {
			t.Fatalf("config adapter: %v", err)
		}
		g.Cfg().SetAdapter(adapter)
		mockDB, mockSQL, err = sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		if err = gdb.Register("sqlmock_guard", &mockDriver{}); err != nil {
			t.Fatalf("register driver: %v", err)
		}
	})
	t.Cleanup(func() {
		if err := mockSQL.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet sqlmock expectations: %v", err)
		}
	})
	return mockSQL
}

// argRecorder 记录绑定参数，并断言金额以参数形式传入。
type argRecorder struct{ got *[]driver.Value }

func (a argRecorder) Match(v driver.Value) bool { *a.got = append(*a.got, v); return true }

func anyArgs(n int, rec *[]driver.Value) []driver.Value {
	out := make([]driver.Value, n)
	for i := range out {
		out[i] = argRecorder{rec}
	}
	return out
}

// 期望的完整 SQL：金额全部为 ? 占位符，SQL 文本中不出现任何数字字面量
const (
	freezeRe   = "^UPDATE `balance_user` SET `balance`=`balance`-\\?,`balance_frozen`=`balance_frozen`\\+\\?,`updated_at`=\\? WHERE \\(`site_id`=\\?\\) AND \\(`user_id`=\\?\\) AND \\(`balance` >= \\?\\)$"
	unfreezeRe = "^UPDATE `balance_user` SET `balance`=`balance`\\+\\?,`balance_frozen`=`balance_frozen`-\\?,`updated_at`=\\? WHERE \\(`site_id`=\\?\\) AND \\(`user_id`=\\?\\) AND \\(`balance_frozen` >= \\?\\)$"
	deductRe   = "^UPDATE `balance_user` SET `balance_frozen`=`balance_frozen`-\\?,`updated_at`=\\? WHERE \\(`site_id`=\\?\\) AND \\(`user_id`=\\?\\) AND \\(`balance_frozen` >= \\?\\)$"
)

func TestFreezeBalance_OK(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := BalanceUser.FreezeBalance(context.Background(), 1, 100, 12.3456); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	// 金额应以参数形式绑定 3 次（balance-?、balance_frozen+?、guard >= ?），且值为正数
	var amt int
	for _, a := range args {
		if f, ok := a.(float64); ok && f == 12.3456 {
			amt++
		}
	}
	if amt != 3 || len(args) != 6 {
		t.Errorf("unexpected bound args: %#v", args)
	}
}

func TestFreezeBalance_SQLParameterized(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	// 若 SQL 中出现浮点字面量或缺少 guard，则正则不匹配，Exec 报错
	mock.ExpectExec(freezeRe).
		WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := BalanceUser.FreezeBalance(context.Background(), 1, 100, 0.1); err != nil {
		t.Fatalf("SQL not parameterized as expected: %v", err)
	}
}

func TestFreezeBalance_InsufficientBalance(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := BalanceUser.FreezeBalance(context.Background(), 1, 100, 999); err == nil {
		t.Fatal("expected error on 0 rows affected (insufficient balance)")
	}
}

func TestGuardedUpdate_NonPositiveAmount(t *testing.T) {
	setupMock(t) // 不设置期望：金额非法时不应发出任何 SQL
	ctx := context.Background()
	for _, amt := range []float64{0, -1, -0.0001} {
		if err := BalanceUser.FreezeBalance(ctx, 1, 100, amt); err == nil {
			t.Errorf("FreezeBalance(%v) expected error", amt)
		}
		if err := BalanceUser.UnfreezeBalance(ctx, 1, 100, amt); err == nil {
			t.Errorf("UnfreezeBalance(%v) expected error", amt)
		}
		if err := BalanceUser.DeductFrozen(ctx, 1, 100, amt); err == nil {
			t.Errorf("DeductFrozen(%v) expected error", amt)
		}
	}
}

func TestFreezeBalance_MissingRecord(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := BalanceUser.FreezeBalance(context.Background(), 1, 404, 1); err == nil {
		t.Fatal("expected error when record missing")
	}
}

func TestUnfreezeBalance_OKAndOverFrozen(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(unfreezeRe).
		WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := BalanceUser.UnfreezeBalance(context.Background(), 1, 100, 5); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	mock.ExpectExec(unfreezeRe).
		WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := BalanceUser.UnfreezeBalance(context.Background(), 1, 100, 1e6); err == nil {
		t.Fatal("expected error when unfreezing more than frozen")
	}
}

func TestDeductFrozen_OKAndOverFrozen(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(deductRe).WithArgs(anyArgs(5, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := BalanceUser.DeductFrozen(context.Background(), 1, 100, 5); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	mock.ExpectExec(deductRe).WithArgs(anyArgs(5, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := BalanceUser.DeductFrozen(context.Background(), 1, 100, 1e6); err == nil {
		t.Fatal("expected error when deducting more than frozen")
	}
}
