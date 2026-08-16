package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testApp struct {
	*App
	t    *testing.T
	out  *bytes.Buffer
	err  *bytes.Buffer
	home string
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	app := &testApp{
		App:  &App{Out: out, Err: errOut, Now: time.Now},
		t:    t,
		out:  out,
		err:  errOut,
		home: testHome(t),
	}

	t.Cleanup(func() {
		if err := app.close(); err != nil {
			t.Errorf("expected no error but got %v, the store is still open", err)
		}
	})

	return app
}

func (a *testApp) run(args ...string) int {
	// Context from t so a command that hangs fails the run instead of stalling
	return a.App.run(a.t.Context(), args)
}

// testHome returns a home directory the test owns. Nothing here calls t.Parallel
// because t.Setenv doesn't work with it.
// TODO: drop the env vars once a flag can point the store somewhere else
func testHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	// os.UserHomeDir reads USERPROFILE on Windows and HOME everywhere else.
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	return home
}

func TestAppRun(t *testing.T) {
	t.Run("only the config hook runs for a config-only command", func(t *testing.T) {
		app := newTestApp(t)

		code := app.run("config", "path")
		if code != 0 {
			t.Fatalf("expected exit code 0 but got %d (output %q)", code, app.out)
		}

		if app.Cfg == nil {
			t.Error("expected a config but got nil, the config hook did not run")
		}

		if app.Store != nil {
			t.Error("expected no store but got one, the db hook ran for a command that doesn't need it")
		}

		want := filepath.Join(app.home, ".config", "transit", "config.yml") + "\n"
		if app.out.String() != want {
			t.Errorf("expected %q but got %q", want, app.out)
		}
	})

	t.Run("both hooks run when the command needs the store", func(t *testing.T) {
		app := newTestApp(t)

		code := app.run("config", "set", "core.location", "dmv", "--verbose")
		if code != 0 {
			t.Fatalf("expected exit code 0 but got %d (output %q)", code, app.out)
		}

		if app.Cfg == nil {
			t.Error("expected a config but got nil, the config hook did not run")
		}

		if app.Store == nil {
			t.Error("expected a store but got nil, the db hook did not run")
		}

		if !app.verbose {
			t.Error("expected verbose to be set, --verbose is not bound to the field")
		}
	})

	t.Run("a failing command writes its diagnostic to Err, not Out", func(t *testing.T) {
		app := newTestApp(t)

		code := app.run("at", "--nonsense")
		if code != 2 {
			t.Fatalf("expected exit code 2 but got %d", code)
		}

		if app.out.Len() != 0 {
			t.Errorf("expected nothing on Out but got %q, a diagnostic reached stdout", app.out)
		}

		if !strings.Contains(app.err.String(), "unknown flag") {
			t.Errorf("expected the flag error on Err but got %q", app.err)
		}
	})
}

func TestAppClose(t *testing.T) {
	t.Run("releases an open store", func(t *testing.T) {
		app := newTestApp(t)

		if code := app.run("config", "set", "core.location", "dmv"); code != 0 {
			t.Fatalf("expected exit code 0 but got %d (output %q)", code, app.out)
		}

		if app.Store == nil {
			t.Fatal("expected a store but got nil, the db hook did not run")
		}

		if err := app.close(); err != nil {
			t.Fatalf("expected no error but got %v", err)
		}

		if err := app.Store.Ping(t.Context()); err == nil {
			t.Error("expected an error pinging a closed store but got nil, the handle is still open")
		}
	})

	t.Run("tolerates a store that was never opened", func(t *testing.T) {
		app := newTestApp(t)

		if code := app.run("config", "path"); code != 0 {
			t.Fatalf("expected exit code 0 but got %d (output %q)", code, app.out)
		}

		if app.Store != nil {
			t.Fatal("expected no store but got one, the db hook ran for a command that doesn't need it")
		}

		if err := app.close(); err != nil {
			t.Errorf("expected no error but got %v", err)
		}
	})
}
