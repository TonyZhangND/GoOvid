package configs

import (
	"testing"

	c "github.com/TonyZhangND/GoOvid/commons"
	a "github.com/TonyZhangND/GoOvid/server/agents"
)

// Tests the correctness of the parser on chat.json
func TestParser_Chat(t *testing.T) {
	res := *Parse("chat.json")

	// check agent 10
	agent := *res[c.ProcessID(10)]
	if agent.Type != a.Chat {
		t.Errorf("agent 10 has type %d; want %d", agent.Type, a.Chat)
	}
	if agent.Box != "127.0.0.1:5000" {
		t.Errorf("agent 10 has box %s; want %s", agent.Box, "127.0.0.1:5000")
	}
	for k, v := range agent.RawAttrs {
		switch k {
		case "myname":
			if v.(string) != "client1" {
				t.Errorf("agent 10 has name %v; want %s", v, "client1")
			}
		case "contacts":
			if len(v.([]interface{})) != 1 {
				t.Errorf("agent 10 has contact %v; want %s", v, "[20]")
			}
			if int(v.([]interface{})[0].(float64)) != 20 {
				t.Errorf("agent 10 has contact %v; want %s", v, "[20]")
			}
		default:
			t.Errorf("agent 10 has invalid attr %s", k)
		}
	}
	for vid, rt := range agent.Routes {
		switch int(vid) {
		case 20:
			if int(rt.DestID) != 20 || int(rt.DestPort) != 100 {
				t.Errorf("agent 10 has invalid route %v, %v", vid, rt)
			}
		default:
			t.Errorf("agent 10 has invalid route %v, %v", vid, rt)
		}
	}

	// check agent 20
	agent = *res[c.ProcessID(20)]
	for k, v := range agent.RawAttrs {
		switch k {
		case "myname":
			if v.(string) != "client2" {
				t.Errorf("agent 20 has name %v; want %s", v, "client2")
			}
		case "contacts":
			if len(v.([]interface{})) != 1 {
				t.Errorf("agent 20 has contact %v; want %s", v, "[10]")
			}
			if int(v.([]interface{})[0].(float64)) != 10 {
				t.Errorf("agent 20 has contact %v; want %s", v, "[10]")
			}
		default:
			t.Errorf("agent 20 has invalid attr %s", k)
		}
	}
}
