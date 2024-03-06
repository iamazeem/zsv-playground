package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type ZsvFlag struct {
	Flag     string
	Argument string
}

type ZsvCommand struct {
	Flags       []ZsvFlag
	SubCommands []string
}

const ()

var ()

func getZsvCommands(version string) bool {
	log.Printf("getting zsv commands [%v]\n", version)

	zsv := generateZsvExePath(version)
	out, err := exec.Command(zsv, "help").Output()
	if err != nil {
		log.Printf("%v\n", err)
		return false
	}

	zsvHelpCommand, ok := parseZsvHelpCommand(string(out))
	if !ok {
		log.Printf("failed to parse 'zsv help' command\n")
		return false
	}

	fmt.Printf("%v\n", zsvHelpCommand)
	return true
}

func parseZsvCommandFlags(scanner *bufio.Scanner) []string {
	flags := []string{}
	for scanner.Scan() {
		flag := strings.TrimSpace(scanner.Text())
		if len(flag) == 0 {
			break
		} else if strings.HasPrefix(flag, "-") {
			flag = strings.TrimSpace(flag[:strings.Index(flag, ":")])
			flags = append(flags, flag)
		}
	}
	return flags
}

func parseZsvCommands(scanner *bufio.Scanner) []string {
	commands := []string{}
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		if len(cmd) == 0 {
			break
		} else if strings.Contains(cmd, ":") {
			cmd = strings.TrimSpace(cmd[:strings.Index(cmd, ":")])
			commands = append(commands, cmd)
		}
	}
	return commands
}

func normalizeZsvFlags(flags []string) []ZsvFlag {
	zsvFlags := []ZsvFlag{}
	for _, flag := range flags {
		index := strings.Index(flag, " ")
		zsvFlg := ZsvFlag{}
		if index == -1 { // without argument
			zsvFlg.Flag = flag
		} else { // with argument
			zsvFlg.Flag = flag[:index]
			zsvFlg.Argument = strings.TrimSpace(flag[index+1:])
		}
		zsvFlags = append(zsvFlags, zsvFlg)
	}
	return zsvFlags
}

func parseZsvHelpCommand(help string) (ZsvCommand, bool) {
	log.Printf("parsing command [zsv help]\n")

	flags := []string{}
	commands := []string{}
	scanner := bufio.NewScanner(strings.NewReader(help))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Options common to all commands") {
			flags = parseZsvCommandFlags(scanner)
		} else if strings.HasPrefix(line, "Commands that parse CSV") {
			commands = parseZsvCommands(scanner)
		}
		if len(flags) > 0 && len(commands) > 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("failed to parse command [zsv help], %v\n", err)
		return ZsvCommand{}, false
	}

	zsvFlags := normalizeZsvFlags(flags)

	log.Printf("parsed command successfully [zsv help]\n")
	return ZsvCommand{Flags: zsvFlags, SubCommands: commands}, true
}

func parseZsvSubcommand(subcommand string) (ZsvCommand, bool) {
	log.Printf("parsing 'zsv %v' subcommand\n", subcommand)

	log.Printf("parsed 'zsv %v' subcommand successfully\n", subcommand)
	return ZsvCommand{}, true
}
