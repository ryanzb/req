package req

import "testing"

func TestGet(t *testing.T) {
	resp, err := Get("https://bsc.for.tube/api/v1/bank_markets?mode=extended")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp.Text())
}

func TestPost(t *testing.T) {
	params := Params{
		"id":      1,
		"jsonrpc": "2.0",
		"params":  List{"f021961"},
		"method":  "filscan.ActorById"}
	resp, err := Post("https://api.filscan.io:8700/rpc/v1", params)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(resp.Text())
}
