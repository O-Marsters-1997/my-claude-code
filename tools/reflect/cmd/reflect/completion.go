package main

import (
	"errors"
	"fmt"
	"strings"
)

var commands = []string{"scan", "slice", "replay", "metrics", "status", "marker", "reports", "uninstall-legacy", "hook", "completion"}

type markerCommand struct {
	name  string
	flags []string
}

var markerCommands = []markerCommand{
	{"harvest", nil},
	{"pull", nil},
	{"add", []string{"--area"}},
	{"ls", []string{"--json", "--status"}},
	{"mark", []string{"--issue", "--proposal", "--scope", "--skill"}},
	{"status", []string{"--remind"}},
}

func markerNames() []string {
	names := make([]string, len(markerCommands))
	for i, m := range markerCommands {
		names[i] = m.name
	}
	return names
}

func completionScript(args []string) (string, error) {
	if len(args) == 1 {
		switch args[0] {
		case "zsh":
			return zshCompletion(), nil
		case "bash":
			return bashCompletion(), nil
		}
	}
	return "", errors.New("usage: reflect completion zsh|bash")
}

func zshCompletion() string {
	var sb strings.Builder
	sb.WriteString("(( $+functions[compdef] )) || { autoload -Uz compinit; compinit; }\n")
	sb.WriteString("_reflect() {\n")
	sb.WriteString("  if (( CURRENT == 2 )); then\n    compadd -- " + strings.Join(commands, " ") + "\n    return\n  fi\n")
	sb.WriteString("  [[ $words[2] == marker ]] || return\n")
	sb.WriteString("  if (( CURRENT == 3 )); then\n    compadd -- " + strings.Join(markerNames(), " ") + "\n    return\n  fi\n")
	sb.WriteString("  if [[ $words[CURRENT-1] == --area ]]; then\n    compadd -U -- ${(f)\"$(reflect complete-area \"$PREFIX\" 2>/dev/null)\"}\n    return\n  fi\n")
	sb.WriteString("  case $words[3] in\n")
	for _, m := range markerCommands {
		if len(m.flags) > 0 {
			fmt.Fprintf(&sb, "    %s) compadd -- %s ;;\n", m.name, strings.Join(m.flags, " "))
		}
	}
	sb.WriteString("  esac\n}\ncompdef _reflect reflect\n")
	return sb.String()
}

func bashCompletion() string {
	var sb strings.Builder
	sb.WriteString("_reflect() {\n  local cur=${COMP_WORDS[COMP_CWORD]} words=\"\"\n")
	sb.WriteString("  if [ \"$COMP_CWORD\" -eq 1 ]; then\n    words=\"" + strings.Join(commands, " ") + "\"\n")
	sb.WriteString("  elif [ \"${COMP_WORDS[1]}\" = marker ] && [ \"${COMP_WORDS[COMP_CWORD-1]}\" = --area ]; then\n")
	sb.WriteString("    local IFS=$'\\n'\n    COMPREPLY=($(reflect complete-area \"$cur\" 2>/dev/null))\n    return\n")
	sb.WriteString("  elif [ \"${COMP_WORDS[1]}\" = marker ]; then\n")
	sb.WriteString("    if [ \"$COMP_CWORD\" -eq 2 ]; then\n      words=\"" + strings.Join(markerNames(), " ") + "\"\n    else\n      case ${COMP_WORDS[2]} in\n")
	for _, m := range markerCommands {
		if len(m.flags) > 0 {
			fmt.Fprintf(&sb, "        %s) words=\"%s\" ;;\n", m.name, strings.Join(m.flags, " "))
		}
	}
	sb.WriteString("      esac\n    fi\n  fi\n")
	sb.WriteString("  COMPREPLY=($(compgen -W \"$words\" -- \"$cur\"))\n}\ncomplete -F _reflect reflect\n")
	return sb.String()
}
