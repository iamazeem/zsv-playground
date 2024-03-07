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
	output, err := exec.Command(zsv, "help").Output()
	if err != nil {
		log.Printf("%v\n", err)
		return false
	}

	zsvHelpCommand, ok := parseZsvHelpCommand(string(output))
	if !ok {
		log.Printf("failed to parse 'zsv help' command\n")
		return false
	}

	fmt.Printf("%v\n", zsvHelpCommand)

	fmt.Printf("global flags\n")
	for _, f := range zsvHelpCommand.Flags {
		fmt.Printf("%v\n", f)
	}

	subcommands := map[string][]ZsvFlag{}

	fmt.Printf("\ncommands\n")
	for _, subcommand := range zsvHelpCommand.SubCommands {
		fmt.Printf("command: %v\n", subcommand)
		output, err := exec.Command(zsv, "help", subcommand).Output()
		if err != nil {
			fmt.Printf("command: %v, error: %v\n", subcommand, err)
			// return false
		}
		subcommands[subcommand] = []ZsvFlag{}
		flags := []string{}
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Options") {
				flags = parseZsvCommandFlags(scanner)
			}
		}
		if len(flags) > 0 {
			subcommands[subcommand] = normalizeZsvCommandFlags(flags)
		}
	}

	fmt.Println(subcommands)

	for subcommand, zsvFlags := range subcommands {
		fmt.Println()
		fmt.Println(subcommand)
		for _, zf := range zsvFlags {
			fmt.Print(zf.Flag)
			if zf.Argument != "" {
				fmt.Println(" |", zf.Argument)
			} else {
				fmt.Println()
			}
		}
		fmt.Println()
	}

	return true
}

func parseZsvCommandFlags(scanner *bufio.Scanner) []string {
	flags := []string{}
	for scanner.Scan() {
		flag := scanner.Text()
		if len(flag) > 3 && strings.HasPrefix(flag, "  -") {
			flag = strings.TrimSpace(flag)
			if !strings.HasPrefix(flag, "-o") {
				index := strings.Index(flag, ":")
				if index != -1 {
					flag = strings.TrimSpace(flag[:strings.Index(flag, ":")])
					flags = append(flags, flag)
				}
			}
		} else if strings.TrimSpace(flag) == "" {
			break
		}
	}
	return flags
}

func normalizeZsvCommandFlags(flags []string) []ZsvFlag {
	zsvFlags := []ZsvFlag{}
	for _, flag := range flags {
		flag = strings.Replace(flag, ", ", ",", 1)
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

func parseZsvMainCommands(scanner *bufio.Scanner) []string {
	commands := []string{}
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			break
		} else if strings.Contains(cmd, ":") {
			cmd = strings.TrimSpace(cmd[:strings.Index(cmd, ":")])
			commands = append(commands, cmd)
		}
	}
	return commands
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
			commands = parseZsvMainCommands(scanner)
		}
		if len(flags) > 0 && len(commands) > 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("failed to parse command [zsv help], %v\n", err)
		return ZsvCommand{}, false
	}

	zsvFlags := normalizeZsvCommandFlags(flags)

	log.Printf("parsed command successfully [zsv help]\n")
	return ZsvCommand{Flags: zsvFlags, SubCommands: commands}, true
}

func parseZsvSubcommand(subcommand string) (ZsvCommand, bool) {
	log.Printf("parsing 'zsv %v' subcommand\n", subcommand)

	log.Printf("parsed 'zsv %v' subcommand successfully\n", subcommand)
	return ZsvCommand{}, true
}
