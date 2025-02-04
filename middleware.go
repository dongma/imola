package imola

// Middleware 函数式的责任链模式
type Middleware func(next HandleFunc) HandleFunc

/*type Net struct {
	handlers []HandleFuncV1
}

func (c Net) Run(ctx *Context) {
	var wg sync.WaitGroup
	for _, hdl := range c.handlers {
		h := hdl
		if h.concurrent {
			wg.Add(1)
			go func() {
				h.Run(ctx)
				wg.Done()
			}()
		} else {
			h.Run(ctx)
		}
	}
	wg.Wait()
}

type HandleFuncV1 struct {
	concurrent bool
	handlers   []*HandleFuncV1
}

func (HandleFuncV1) Run(ctx *Context) {

}*/
