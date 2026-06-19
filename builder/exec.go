package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// execOpts holds the resolved execution settings derived from a Spec.
type execOpts struct {
	bin    string
	dryRun bool
	stdout io.Writer
	stderr io.Writer
}

func (s Spec) execOpts() execOpts {
	o := execOpts{bin: s.Bin, dryRun: s.DryRun, stdout: s.Stdout, stderr: s.Stderr}
	if o.bin == "" {
		o.bin = "docker"
	}
	if o.stdout == nil {
		o.stdout = os.Stdout
	}
	if o.stderr == nil {
		o.stderr = os.Stderr
	}
	return o
}

// run executes the binary with args, or prints the command when dryRun is set.
func (o execOpts) run(ctx context.Context, args []string) error {
	if o.dryRun {
		fmt.Fprintln(o.stdout, shellCommand(o.bin, args))
		return nil
	}

	cmd := exec.CommandContext(ctx, o.bin, args...)
	cmd.Stdout = o.stdout
	cmd.Stderr = o.stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", o.bin, strings.Join(args, " "), err)
	}
	return nil
}

// shellCommand renders a copy-pasteable command line, quoting arguments that
// contain shell-significant characters.
func shellCommand(bin string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, shellQuote(bin))
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if strings.ContainsAny(s, " \t\n\"'$&|;<>()*?[]{}#`\\!~") {
		return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
	}
	return s
}
