package main

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
)

type ZsvFlag struct {
	Flag     string
	Argument string
}

type ZsvCommand struct {
	Flags    []ZsvFlag
	Commands []string
}

// command -> flags
type ZsvCommands map[string][]ZsvFlag

// version -> commands
type zsvCLI map[string]ZsvCommands

func loadCLIForAllZsvVersions(versions []string) (zsvCLI, bool) {
	log.Printf("loading CLI for all zsv versions [%v]", versions)

	zsv := zsvCLI{}
	for _, version := range versions {
		zsvCommands, ok := loadZsvCommands(version)
		if !ok {
			log.Printf("failed load CLI for zsv %v", version)
			return nil, false
		}
		zsv[version] = zsvCommands
	}

	log.Printf("loading CLI for all zsv versions successfully [%v]", zsv)
	return zsv, true
}

func loadZsvCommands(version string) (ZsvCommands, bool) {
	log.Printf("loading zsv commands [%v]", version)

	zsv := getZsvExePath(version)
	zsvHelpCommand, ok := loadZsvHelpCommand(zsv)
	if !ok {
		log.Print("failed to parse 'zsv help' command")
		return nil, false
	}

	log.Print("listing global flags")
	for _, zsvFlag := range zsvHelpCommand.Flags {
		if zsvFlag.Argument == "" {
			log.Print(zsvFlag.Flag)
		} else {
			log.Print(zsvFlag.Flag, " | ", zsvFlag.Argument)
		}
	}

	subcommands := loadZsvSubcommands(zsv, zsvHelpCommand)
	log.Printf("loaded zsv commands successfully [%v]", version)
	return subcommands, true
}

func parseZsvCommandFlags(scanner *bufio.Scanner) []string {
	flags := []string{}
	for scanner.Scan() {
		flag := scanner.Text()
		if len(flag) > 3 && strings.HasPrefix(flag, "  -") {
			flag = strings.TrimSpace(flag)
			if !strings.HasPrefix(flag, "-o") && !strings.HasPrefix(flag, "-h") {
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
			zsvFlg.Argument = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(flag[index+1:])), " ", "_")
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

func loadZsvHelpCommand(zsv string) (ZsvCommand, bool) {
	log.Print("loading 'zsv help' command")

	output, err := exec.Command(zsv, "help").Output()
	if err != nil {
		log.Printf("failed to get output of 'zsv help', error: %v", err)
		return ZsvCommand{}, false
	}

	flags := []string{}
	commands := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
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
		log.Printf("failed to parse 'zsv help' command, error: %v", err)
		return ZsvCommand{}, false
	}

	zsvFlags := normalizeZsvCommandFlags(flags)
	zsvHelpCommand := ZsvCommand{
		Flags:    zsvFlags,
		Commands: commands,
	}

	log.Printf("zsv help: %v", zsvHelpCommand)
	log.Print("load 'zsv help' command successfully")
	return zsvHelpCommand, true
}

func loadZsvSubcommands(zsv string, zsvHelpCommand ZsvCommand) map[string][]ZsvFlag {
	log.Print("loading all subcommands with flags")

	subcommands := map[string][]ZsvFlag{}

	for _, subcommand := range zsvHelpCommand.Commands {
		log.Printf("subcommand: %v", subcommand)
		output, err := exec.Command(zsv, "help", subcommand).Output()
		if err != nil {
			log.Printf("subcommand: %v, error: %v", subcommand, err)
			// zsv help 2json returns exit code 5
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

	log.Print("listing subcommands with flags: ", subcommands)
	for subcommand, zsvFlags := range subcommands {
		log.Print("subcommand: ", subcommand)
		for _, zsvFlag := range zsvFlags {
			if zsvFlag.Argument == "" {
				log.Print(zsvFlag.Flag)
			} else {
				log.Print(zsvFlag.Flag, " | ", zsvFlag.Argument)
			}
		}
	}

	log.Print("loaded all subcommands with flags successfully")
	return subcommands
}
