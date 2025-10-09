package lua_modules

import lua "github.com/yuin/gopher-lua"

// работа с http
type HttpClient interface {
	Get(url string, headers map[string]string) ([]byte, int, error)
	Post(url string, body []byte, headers map[string]string) ([]byte, int, error)
	Fetch(method, url string, body []byte, headers map[string]string) ([]byte, int, error)
}

func (m *HttpModule) applyGet(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		headersTable := L.OptTable(2, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		body, status, err := m.client.Get(url, headers)
		return pushHttpResponse(L, body, status, err)
	}))
}

func (m *HttpModule) applyPost(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		url := L.CheckString(1)
		body := L.CheckString(2)
		headersTable := L.OptTable(3, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		respBody, status, err := m.client.Post(url, []byte(body), headers)
		return pushHttpResponse(L, respBody, status, err)
	}))
}

func (m *HttpModule) applyRequest(L *lua.LState, cmd string) {
	L.SetGlobal(cmd, L.NewFunction(func(L *lua.LState) int {
		method := L.CheckString(1)
		url := L.CheckString(2)
		body := L.CheckString(3)
		headersTable := L.OptTable(4, L.NewTable())

		headers := make(map[string]string)
		headersTable.ForEach(func(k, v lua.LValue) {
			headers[k.String()] = v.String()
		})

		respBody, status, err := m.client.Fetch(method, url, []byte(body), headers)
		return pushHttpResponse(L, respBody, status, err)
	}))
}

func pushHttpResponse(L *lua.LState, body []byte, status int, err error) int {
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	tbl := L.NewTable()
	tbl.RawSetString("status", lua.LNumber(status))
	tbl.RawSetString("body", lua.LString(string(body)))
	L.Push(tbl)
	return 1
}

type HttpModule struct{ client HttpClient }

func NewHttp(client HttpClient) *HttpModule {
	return &HttpModule{client: client}
}
