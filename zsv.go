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

	// log.Printf("got zsv commands [%v]\n", commands)
	return true
}

func parseZsvHelpCommand(help string) (ZsvCommand, bool) {
	log.Printf("parsing help command\n")

	flags := []string{}
	commands := []string{}
	scanner := bufio.NewScanner(strings.NewReader(help))
	for scanner.Scan() {
		line := scanner.Text()
		// global flags
		if strings.HasPrefix(line, "Options common to all commands") {
			for scanner.Scan() {
				flag := strings.TrimSpace(scanner.Text())
				if len(flag) == 0 {
					break
				} else if strings.HasPrefix(flag, "-") {
					flag = strings.TrimSpace(flag[:strings.Index(flag, ":")])
					flags = append(flags, flag)
				}
			}
		}
		// main commands
		if strings.HasPrefix(line, "Commands that parse CSV") {
			for scanner.Scan() {
				cmd := strings.TrimSpace(scanner.Text())
				if len(cmd) == 0 {
					break
				} else if strings.Contains(cmd, ":") {
					cmd = strings.TrimSpace(cmd[:strings.Index(cmd, ":")])
					commands = append(commands, cmd)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("failed to parse, %v\n", err)
		return ZsvCommand{}, false
	}

	zsvFlags := parseZsvFlags(flags)

	log.Printf("parsed help command successfully\n")
	return ZsvCommand{Flags: zsvFlags, SubCommands: commands}, true
}

func parseZsvSubcommand(subcommand string) (ZsvCommand, bool) {
	log.Printf("parsing 'zsv %v' subcommand\n", subcommand)

	log.Printf("parsed 'zsv %v' subcommand successfully\n", subcommand)
	return ZsvCommand{}, true
}

func parseZsvFlags(flags []string) []ZsvFlag {
	zsvFlags := []ZsvFlag{}
	for _, flag := range flags {
		index := strings.Index(flag, " ")
		zf := ZsvFlag{}
		if index == -1 { // without argument
			zf.Flag = flag
		} else { // with argument
			zf.Flag = flag[:index]
			zf.Argument = flag[index+1:]
		}
		zsvFlags = append(zsvFlags, zf)
	}
	return zsvFlags
}
