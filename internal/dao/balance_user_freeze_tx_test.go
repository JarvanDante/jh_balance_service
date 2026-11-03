package dao

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/frame/g"
)

// 复用 balance_user_guard_test.go 中的 setupMock / freezeRe / anyArgs。

func TestFreezeBalanceTx_RunsOnTxAndCommits(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectBegin()
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	ctx := context.Background()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := BalanceUser.FreezeBalanceTx(ctx, tx, 1, 100, 8.5); err != nil {
		_ = tx.Rollback()
		t.Fatalf("expected success, got %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	var amt int
	for _, a := range args {
		if f, ok := a.(float64); ok && f == 8.5 {
			amt++
		}
	}
	if amt != 3 {
		t.Errorf("amount should be bound 3 times, args=%#v", args)
	}
}

func TestFreezeBalanceTx_InsufficientRollsBack(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectBegin()
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	ctx := context.Background()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := BalanceUser.FreezeBalanceTx(ctx, tx, 1, 100, 1e6); err == nil {
		t.Error("expected error when balance insufficient")
	}
	_ = tx.Rollback()
}

func TestFreezeBalanceTx_MissingRecord(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectBegin()
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	ctx := context.Background()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := BalanceUser.FreezeBalanceTx(ctx, tx, 1, 404, 1); err == nil {
		t.Error("expected error when record missing")
	}
	_ = tx.Rollback()
}

func TestFreezeBalanceTx_NonPositiveAmountNoSQL(t *testing.T) {
	mock := setupMock(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	ctx := context.Background()
	tx, err := g.DB().Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for _, amt := range []float64{0, -3} {
		if err := BalanceUser.FreezeBalanceTx(ctx, tx, 1, 100, amt); err == nil {
			t.Errorf("FreezeBalanceTx(%v) expected error", amt)
		}
	}
	_ = tx.Rollback()
}

func TestFreezeBalanceTx_NilTxFallsBackToNoTx(t *testing.T) {
	mock := setupMock(t)
	var args []driver.Value
	mock.ExpectExec(freezeRe).WithArgs(anyArgs(6, &args)...).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := BalanceUser.FreezeBalanceTx(context.Background(), nil, 1, 100, 2); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}
