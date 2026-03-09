package contextcases

import "context"

func take(ctx context.Context) {}

func badBackground(ctx context.Context) {
	take(context.Background()) // want "avoid context.Background/TODO"
}

func badAssignedBackground(ctx context.Context) {
	bgCtx := context.Background() // want "avoid context.Background/TODO"
	take(bgCtx)                   // want "pass ctx or a context derived from ctx"
}

func badTodoInGoroutine(ctx context.Context) {
	go func() {
		take(context.TODO()) // want "avoid context.Background/TODO"
	}()
}

func goodDerived(ctx context.Context) {
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	take(childCtx)
}

func goodNoCtx() {
	take(context.Background())
}
