// Bank: custom errors, errors.Is / errors.As, wrapping, panic and recover.
package main

import (
	"errors"
	"fmt"
	"strings"
)

var ErrFrozen = errors.New("account frozen")

type InsufficientFunds struct {
	Need, Have int
}

func (e *InsufficientFunds) Error() string {
	return fmt.Sprintf("insufficient funds: need %d, have %d", e.Need, e.Have)
}

type Account struct {
	ID      string
	Balance int
	Frozen  bool
	History []string
}

func (a *Account) log(format string, args ...any) {
	a.History = append(a.History, fmt.Sprintf(format, args...))
}

func (a *Account) Deposit(n int) error {
	if a.Frozen {
		return fmt.Errorf("deposit to %s: %w", a.ID, ErrFrozen)
	}
	if n <= 0 {
		panic(fmt.Sprintf("invalid deposit %d", n))
	}
	a.Balance += n
	a.log("deposit %d", n)
	return nil
}

func (a *Account) Withdraw(n int) error {
	if a.Frozen {
		return fmt.Errorf("withdraw from %s: %w", a.ID, ErrFrozen)
	}
	if a.Balance < n {
		return fmt.Errorf("withdraw from %s: %w", a.ID, &InsufficientFunds{n, a.Balance})
	}
	a.Balance -= n
	a.log("withdraw %d", n)
	return nil
}

type Bank struct {
	accounts map[string]*Account
}

func (b *Bank) Open(id string, initial int) *Account {
	acc := &Account{ID: id, Balance: initial}
	if b.accounts == nil {
		b.accounts = map[string]*Account{}
	}
	b.accounts[id] = acc
	return acc
}

func (b *Bank) Transfer(from, to string, n int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("transfer aborted: %v", r)
		}
	}()
	src, ok := b.accounts[from]
	if !ok {
		return fmt.Errorf("no account %q", from)
	}
	dst, ok := b.accounts[to]
	if !ok {
		return fmt.Errorf("no account %q", to)
	}
	if err := src.Withdraw(n); err != nil {
		return err
	}
	if err := dst.Deposit(n); err != nil {
		src.Balance += n
		src.log("rollback %d", n)
		return err
	}
	return nil
}

func classify(err error) string {
	var short *InsufficientFunds
	switch {
	case err == nil:
		return "ok"
	case errors.Is(err, ErrFrozen):
		return "frozen"
	case errors.As(err, &short):
		return fmt.Sprintf("short by %d", short.Need-short.Have)
	default:
		return "other"
	}
}

func main() {
	bank := &Bank{}
	bank.Open("alice", 100)
	bank.Open("bob", 20)
	frozen := bank.Open("carol", 500)
	frozen.Frozen = true

	tests := []struct {
		from, to string
		amount   int
	}{
		{"alice", "bob", 30},
		{"bob", "alice", 80},
		{"alice", "carol", 10},
		{"dave", "bob", 1},
		{"alice", "bob", -5},
		{"carol", "bob", 5},
	}
	for i, t := range tests {
		err := bank.Transfer(t.from, t.to, t.amount)
		fmt.Printf("%d. %s -> %s %d: %-14s %v\n", i+1, t.from, t.to, t.amount, classify(err), err)
	}
	for _, id := range []string{"alice", "bob", "carol"} {
		acc := bank.accounts[id]
		fmt.Printf("%-6s %4d  [%s]\n", id, acc.Balance, strings.Join(acc.History, "; "))
	}
	joined := errors.Join(ErrFrozen, &InsufficientFunds{5, 1})
	fmt.Println(errors.Is(joined, ErrFrozen), strings.ReplaceAll(joined.Error(), "\n", " | "))
}
