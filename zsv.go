package main

import (
	"bufio"
	"log"
	"os/exec"
	"strings"
)

type ZsvFlag struct {
	Name     string `json:"name"`
	Argument string `json:"argument,omitempty"`
}

type ZsvCommand struct {
	Name  string    `json:"name"`
	Flags []ZsvFlag `json:"flags,omitempty"`
}

type ZsvCLI struct {
	GlobalFlags []ZsvFlag    `json:"global_flags"`
	Commands    []ZsvCommand `json:"commands"`
}

// version -> CLI
type ZsvCLIs map[string]ZsvCLI

func loadCLIs(versions []string) (ZsvCLIs, bool) {
	log.Printf("loading CLIs for all zsv versions %v", versions)

	clis := ZsvCLIs{}
	for _, version := range versions {
		cli, ok := loadCLI(version)
		if !ok {
			log.Printf("failed to load CLI for zsv %v", version)
			return nil, false
		}
		clis[version] = cli
	}

	log.Printf("loaded CLIs for all zsv versions successfully")
	return clis, true
}

func loadCLI(version string) (ZsvCLI, bool) {
	log.Printf("loading CLI [%v]", version)

	zsv := getZsvExePath(version)
	globalFlags, commandList, ok := loadGlobalFlagsAndCommands(zsv)
	if !ok {
		log.Print("failed to parse 'zsv help' command")
		return ZsvCLI{}, false
	}

	// log.Print("listing global flags")
	// for _, zsvFlag := range globalFlags {
	// 	if zsvFlag.Argument == "" {
	// 		log.Print(zsvFlag.Name)
	// 	} else {
	// 		log.Printf("%v | %v", zsvFlag.Name, zsvFlag.Argument)
	// 	}
	// }

	commands := loadCommands(zsv, commandList)
	log.Printf("loaded CLI successfully [%v]", version)
	return ZsvCLI{GlobalFlags: globalFlags, Commands: commands}, true
}

func parseFlags(scanner *bufio.Scanner) []ZsvFlag {
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
	return normalizeFlags(flags)
}

func normalizeFlags(flags []string) []ZsvFlag {
	zsvFlags := []ZsvFlag{}
	for _, flag := range flags {
		flag = strings.Replace(flag, ", ", ",", 1)
		index := strings.Index(flag, " ")
		zsvFlg := ZsvFlag{}
		if index == -1 { // without argument
			zsvFlg.Name = flag
		} else { // with argument
			zsvFlg.Name = flag[:index]
			zsvFlg.Argument = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(flag[index+1:])), " ", "_")
		}
		zsvFlags = append(zsvFlags, zsvFlg)
	}
	return zsvFlags
}

func parseCommands(scanner *bufio.Scanner) []string {
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

func loadGlobalFlagsAndCommands(zsv string) ([]ZsvFlag, []string, bool) {
	log.Print("loading global flags and commands")

	output, err := exec.Command(zsv, "help").Output()
	if err != nil {
		log.Printf("failed to get output of 'zsv help', error: %v", err)
		return nil, nil, false
	}

	globalFlags := []ZsvFlag{}
	commands := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Options common to all commands") {
			globalFlags = parseFlags(scanner)
		} else if strings.HasPrefix(line, "Commands that parse CSV") {
			commands = parseCommands(scanner)
		}
		if len(globalFlags) > 0 && len(commands) > 0 {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("failed to parse 'zsv help' command, error: %v", err)
		return nil, nil, false
	}

	log.Printf("global flags: %v", globalFlags)
	log.Printf("commands: %v", commands)

	log.Print("loaded global flags and commands successfully")
	return globalFlags, commands, true
}

func loadCommands(zsv string, commandList []string) []ZsvCommand {
	log.Print("loading all commands")

	commands := []ZsvCommand{}

	for _, command := range commandList {
		// log.Printf("command: %v", command)
		output, err := exec.Command(zsv, "help", command).Output()
		if err != nil {
			log.Printf("command: %v, error: %v", command, err)
			// zsv help 2json returns exit code 5
			// return false
		}
		flags := []ZsvFlag{}
		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Options") {
				flags = parseFlags(scanner)
			}
		}
		commands = append(commands, ZsvCommand{Name: command, Flags: flags})
	}

	// log.Printf("listing commands with flags: %v", commands)
	// for _, command := range commands {
	// 	log.Printf("command: %v", command)
	// 	for _, flag := range command.Flags {
	// 		if flag.Argument == "" {
	// 			log.Print(flag.Name)
	// 		} else {
	// 			log.Print(flag.Name, " | ", flag.Argument)
	// 		}
	// 	}
	// }

	log.Print("loaded all commands successfully")
	return commands
}
