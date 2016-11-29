// Package mock implements checkin.Service and provides testing utilities.
package mock

import (
	"errors"

	"github.com/micromdm/mdm"
	"golang.org/x/net/context"
)

type CheckinService struct {
	AuthenticateInvoked bool
	AuthenticateFunc    CheckinFunc

	TokenUpdateInvoked bool
	TokenUpdateFunc    CheckinFunc

	CheckOutInvoked bool
	CheckoutFunc    CheckinFunc
}

type CheckinFunc func(ctx context.Context, cmd mdm.CheckinCommand) error

func (svc *CheckinService) Authenticate(ctx context.Context, cmd mdm.CheckinCommand) error {
	svc.AuthenticateInvoked = true
	return svc.AuthenticateFunc(ctx, cmd)
}

func (svc *CheckinService) TokenUpdate(ctx context.Context, cmd mdm.CheckinCommand) error {
	svc.TokenUpdateInvoked = true
	return svc.TokenUpdateFunc(ctx, cmd)
}

func (svc *CheckinService) CheckOut(ctx context.Context, cmd mdm.CheckinCommand) error {
	svc.CheckOutInvoked = true
	return svc.CheckoutFunc(ctx, cmd)
}

func FailCheckin(context.Context, mdm.CheckinCommand) error {
	return errors.New("checkin failed")
}
func SucceedCheckin(context.Context, mdm.CheckinCommand) error {
	return nil
}
