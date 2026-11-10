package test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	apiv1 "github.com/usememos/memos/proto/gen/api/v1"
)

func TestMemoHideUntilLifecycle(t *testing.T) {
	ctx := context.Background()
	ts := NewTestService(t)
	defer ts.Cleanup()

	user, err := ts.CreateRegularUser(ctx, "hide-until-owner")
	require.NoError(t, err)
	userCtx := ts.CreateUserContext(ctx, user.ID)
	future := timestamppb.New(time.Now().Add(time.Hour))

	hidden, err := ts.Service.CreateMemo(userCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "future hidden",
		Visibility: apiv1.Visibility_PRIVATE,
		HideUntil:  future,
	}})
	require.NoError(t, err)
	require.Equal(t, future, hidden.HideUntil)

	visible, err := ts.Service.CreateMemo(userCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "visible",
		Visibility: apiv1.Visibility_PRIVATE,
	}})
	require.NoError(t, err)

	feed, err := ts.Service.ListMemos(userCtx, &apiv1.ListMemosRequest{PageSize: 10})
	require.NoError(t, err)
	require.True(t, slices.ContainsFunc(feed.Memos, func(memo *apiv1.Memo) bool { return memo.Name == visible.Name }))
	require.False(t, slices.ContainsFunc(feed.Memos, func(memo *apiv1.Memo) bool { return memo.Name == hidden.Name }))

	direct, err := ts.Service.GetMemo(userCtx, &apiv1.GetMemoRequest{Name: hidden.Name})
	require.NoError(t, err)
	require.Equal(t, future, direct.HideUntil)

	revealed, err := ts.Service.UpdateMemo(userCtx, &apiv1.UpdateMemoRequest{
		Memo:       &apiv1.Memo{Name: hidden.Name},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"hide_until"}},
	})
	require.NoError(t, err)
	require.Nil(t, revealed.HideUntil)

	feed, err = ts.Service.ListMemos(userCtx, &apiv1.ListMemosRequest{PageSize: 10})
	require.NoError(t, err)
	require.True(t, slices.ContainsFunc(feed.Memos, func(memo *apiv1.Memo) bool { return memo.Name == hidden.Name }))
}

func TestMemoHideUntilRejectsInvalidTimestamp(t *testing.T) {
	ctx := context.Background()
	ts := NewTestService(t)
	defer ts.Cleanup()

	user, err := ts.CreateRegularUser(ctx, "invalid-hide-until-owner")
	require.NoError(t, err)
	userCtx := ts.CreateUserContext(ctx, user.ID)
	invalid := &timestamppb.Timestamp{Seconds: 253402300800}

	_, err = ts.Service.CreateMemo(userCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "invalid",
		Visibility: apiv1.Visibility_PRIVATE,
		HideUntil:  invalid,
	}})
	require.Equal(t, codes.InvalidArgument, status.Code(err))

	memo, err := ts.Service.CreateMemo(userCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "valid",
		Visibility: apiv1.Visibility_PRIVATE,
	}})
	require.NoError(t, err)
	_, err = ts.Service.UpdateMemo(userCtx, &apiv1.UpdateMemoRequest{
		Memo:       &apiv1.Memo{Name: memo.Name, HideUntil: invalid},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"hide_until"}},
	})
	require.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestListMemosHiddenOnly(t *testing.T) {
	ctx := context.Background()
	ts := NewTestService(t)
	defer ts.Cleanup()

	owner, err := ts.CreateRegularUser(ctx, "hidden-only-owner")
	require.NoError(t, err)
	ownerCtx := ts.CreateUserContext(ctx, owner.ID)
	other, err := ts.CreateRegularUser(ctx, "hidden-only-other")
	require.NoError(t, err)
	otherCtx := ts.CreateUserContext(ctx, other.ID)
	future := timestamppb.New(time.Now().Add(time.Hour))

	hidden, err := ts.Service.CreateMemo(ownerCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "postponed",
		Visibility: apiv1.Visibility_PUBLIC,
		HideUntil:  future,
	}})
	require.NoError(t, err)
	_, err = ts.Service.CreateMemo(ownerCtx, &apiv1.CreateMemoRequest{Memo: &apiv1.Memo{
		Content:    "visible",
		Visibility: apiv1.Visibility_PUBLIC,
	}})
	require.NoError(t, err)

	postponed, err := ts.Service.ListMemos(ownerCtx, &apiv1.ListMemosRequest{PageSize: 10, HiddenOnly: true})
	require.NoError(t, err)
	require.Len(t, postponed.Memos, 1)
	require.Equal(t, hidden.Name, postponed.Memos[0].Name)

	othersView, err := ts.Service.ListMemos(otherCtx, &apiv1.ListMemosRequest{PageSize: 10, HiddenOnly: true})
	require.NoError(t, err)
	require.Empty(t, othersView.Memos, "the postponed view only lists the caller's own memos")
}
