package bind

import "testing"

func TestPruneKeepsWhatTheRootsReach(t *testing.T) {
	module, _ := ParseDeclarations(`
declare class Engine { constructor(); world(): World; helper(): Helper }
declare class World { add(b: Body): void }
declare class Body { x: number }
declare class Helper { tool(): Tool }
declare class Tool { }
declare class Unused { }
declare function make(): Engine;
declare function other(): Unused;
`)

	if missing := Prune(module, []string{"make"}, 2); len(missing) != 0 {
		t.Fatalf("missing %v", missing)
	}

	have := map[string]bool{}

	for _, c := range module.Classes {
		have[c.Name] = true
	}

	for _, name := range []string{"Engine", "World", "Helper"} {
		if !have[name] {
			t.Errorf("%s was cut", name)
		}
	}

	for _, name := range []string{"Unused", "Body", "Tool"} {
		if have[name] {
			t.Errorf("%s was kept", name)
		}
	}

	if len(module.Functions) != 1 || module.Functions[0].Name != "make" {
		t.Errorf("functions: %+v", module.Functions)
	}

	// what is cut is any where it was mentioned
	for _, c := range module.Classes {
		if c.Name == "World" && c.Methods[0].Params[0].Type.Kind != Any {
			t.Errorf("a type that was cut is %+v", c.Methods[0].Params[0].Type)
		}
	}

	if missing := Prune(module, []string{"nothing"}, 0); len(missing) != 1 {
		t.Errorf("a name that is not there was not told: %v", missing)
	}
}

func TestACallbackInterfaceIsAFunctionAndEventsAreMethods(t *testing.T) {
	module, _ := ParseDeclarations(`
interface Tick { (time: number): void }
interface Key { code: string }
interface PadEventMap { "key": Key }
declare class Pad {
	addEventListener<K extends keyof PadEventMap>(type: K, listener: (this: Pad, ev: PadEventMap[K]) => any): void;
	run(done: Tick | { handleEvent(): void }): void;
}
`)

	for _, c := range module.Classes {
		if c.Name != "Pad" {
			continue
		}

		var on, run *Function

		for i := range c.Methods {
			switch c.Methods[i].Name {
			case "OnKey":
				on = &c.Methods[i]
			case "run":
				run = &c.Methods[i]
			}
		}

		if on == nil || on.JS != "addEventListener" || len(on.Lead) != 1 || on.Lead[0] != `"key"` {
			t.Fatalf("OnKey: %+v", on)
		}

		if listener := on.Params[0].Type; listener.Kind != Func || listener.Params[0].Type.Kind != Struct {
			t.Errorf("the listener is %+v", listener)
		}

		if run == nil || run.Params[0].Type.Kind != Func {
			t.Errorf("a callback or an object that has one is %+v", run)
		}

		return
	}

	t.Fatal("no Pad")
}
