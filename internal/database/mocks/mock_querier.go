package mocks

import (
	"context"

	"github.com/jonathanschwarzhaupt/my-blog/internal/database"
)

type MockQuerier struct {
	GetPostFunc    func(ctx context.Context, slug string) (database.Post, error)
	InsertPostFunc func(ctx context.Context, arg database.InsertPostParams) (database.Post, error)
	ListPostsFunc  func(ctx context.Context) ([]database.Post, error)
	UpdatePostFunc func(ctx context.Context, arg database.UpdatePostParams) (database.Post, error)
}

func (m *MockQuerier) GetPost(ctx context.Context, slug string) (database.Post, error) {
	return m.GetPostFunc(ctx, slug)
}

func (m *MockQuerier) InsertPost(ctx context.Context, arg database.InsertPostParams) (database.Post, error) {
	return m.InsertPostFunc(ctx, arg)
}

func (m *MockQuerier) ListPosts(ctx context.Context) ([]database.Post, error) {
	return m.ListPostsFunc(ctx)
}

func (m *MockQuerier) UpdatePost(ctx context.Context, arg database.UpdatePostParams) (database.Post, error) {
	return m.UpdatePostFunc(ctx, arg)
}

var _ database.Querier = (*MockQuerier)(nil)
