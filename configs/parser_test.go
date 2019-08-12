package configs

import (
	"testing"

	a "github.com/TonyZhangND/GoOvid/agents"
	c "github.com/TonyZhangND/GoOvid/commons"
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
			if v.(string) != "alice" {
				t.Errorf("agent 10 has name %v; want %s", v, "alice")
			}
		case "contacts":
			if len(v.([]interface{})) != 2 {
				t.Errorf("agent 10 has contact %v; want %s", v, "[20, 30]")
			}
			if int(v.([]interface{})[0].(float64)) != 20 {
				t.Errorf("agent 10 has contact %v; want %s", v, "[20, 30]")
			}
			if int(v.([]interface{})[1].(float64)) != 30 {
				t.Errorf("agent 10 has contact %v; want %s", v, "[20, 30]")
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
		case 30:
			if int(rt.DestID) != 30 || int(rt.DestPort) != 100 {
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
			if v.(string) != "bob" {
				t.Errorf("agent 20 has name %v; want %s", v, "bob")
			}
		case "contacts":
			if len(v.([]interface{})) != 2 {
				t.Errorf("agent 20 has contact %v; want %s", v, "[10, 30]")
			}
			if int(v.([]interface{})[0].(float64)) != 10 {
				t.Errorf("agent 20 has contact %v; want %s", v, "[10, 30]")
			}
			if int(v.([]interface{})[1].(float64)) != 30 {
				t.Errorf("agent 20 has contact %v; want %s", v, "[10, 30]")
			}
		default:
			t.Errorf("agent 20 has invalid attr %s", k)
		}
	}
}
