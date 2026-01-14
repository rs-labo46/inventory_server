package stock_test

import (
	"context"
	"errors"
	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
	"testing"
)

type stockRepoStub struct {
	change func(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error)
}

func (s stockRepoStub) Change(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
	return s.change(ctx, actorUserID, productID, delta, reason)
}

var _ repo.StockRepository = stockRepoStub{}

func TestUsecase_Inbound(t *testing.T) {
	t.Parallel()

	t.Run("qty<=0 は ErrInvalid", func(t *testing.T) {
		t.Parallel()

		uc := stock.Usecase{
			Stock: stockRepoStub{
				change: func(ctx context.Context, actorUserID, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
					t.Fatalf("qty<=0 では Change は呼ばれない想定")
					return domain.Inventory{}, domain.StockMovement{}, nil
				},
			},
		}

		_, err := uc.Inbound(context.Background(), "u1", "p1", 0, "test")
		if !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrInvalid)
		}
	})

	t.Run("正常はChangeに+qtyが渡され、結果が変えられる", func(t *testing.T) {
		t.Parallel()

		called := false

		wantActor := "u1"
		wantProduct := "p1"
		wantDelta := 10
		wantReason := "purchase"

		wantInv := domain.Inventory{}
		wantMv := domain.StockMovement{}

		uc := stock.Usecase{
			Stock: stockRepoStub{
				change: func(ctx context.Context, actorUserID, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
					called = true

					if actorUserID != wantActor {
						t.Fatalf("actorUserID mismatch: got=%s want=%s", actorUserID, wantActor)
					}
					if productID != wantProduct {
						t.Fatalf("productID mismatch: got=%s want=%s", productID, wantProduct)
					}
					if delta != wantDelta {
						t.Fatalf("delta mismatch: got=%d want=%d", delta, wantDelta)
					}
					if reason != wantReason {
						t.Fatalf("reason mismatch: got=%s want=%s", reason, wantReason)
					}
					return wantInv, wantMv, nil
				},
			},
		}

		got, err := uc.Inbound(context.Background(), wantActor, wantProduct, wantDelta, wantReason)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("Change was not called")
		}

		if got.Inventory != wantInv {
			t.Fatalf("inventory mismatch")
		}
		if got.Movement != wantMv {
			t.Fatalf("movement mismatch")
		}
	})

	t.Run("Change がエラーを返したらそのまま返す", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("some error")

		uc := stock.Usecase{
			Stock: stockRepoStub{
				change: func(ctx context.Context, actorUserID, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
					return domain.Inventory{}, domain.StockMovement{}, wantErr
				},
			},
		}

		_, err := uc.Inbound(context.Background(), "u1", "p1", 1, "x")
		if !errors.Is(err, wantErr) {
			t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
		}
	})
}

func TestUsecase_Outbound(t *testing.T) {
	t.Parallel()

	t.Run("qty<=0 は ErrInvalid", func(t *testing.T) {
		t.Parallel()

		uc := stock.Usecase{
			Stock: stockRepoStub{
				change: func(ctx context.Context, actorUserID, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
					t.Fatalf("qty<=0 では Change は呼ばれない想定")
					return domain.Inventory{}, domain.StockMovement{}, nil
				},
			},
		}

		_, err := uc.Outbound(context.Background(), "u1", "p1", -1, "test")
		if !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrInvalid)
		}
	})

	t.Run("正常系: Change に -qty が渡される", func(t *testing.T) {
		t.Parallel()

		wantQty := 3
		wantDelta := -3

		uc := stock.Usecase{
			Stock: stockRepoStub{
				change: func(ctx context.Context, actorUserID, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
					if delta != wantDelta {
						t.Fatalf("delta mismatch: got=%d want=%d", delta, wantDelta)
					}
					return domain.Inventory{}, domain.StockMovement{}, nil
				},
			},
		}

		_, err := uc.Outbound(context.Background(), "u1", "p1", wantQty, "ship")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
