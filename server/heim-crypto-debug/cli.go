package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

var (
	spaceRe = regexp.MustCompile("^\\s+")
	wordRe  = regexp.MustCompile("^([^\"\\s]+|\"([^\"\\\\]+|\\\\.)*\")")
)

type fieldDesc struct {
	name       string
	usage      string
	positional bool
	required   bool
}

func newFieldDesc(field reflect.StructField) *fieldDesc {
	tags := field.Tag
	usage, ok := tags.Lookup("usage")
	if !ok {
		return nil
	}
	info, _ := tags.Lookup("cli")
	infoParts := strings.Split(info, ",")
	name := infoParts[0]
	if name == "" {
		for i, c := range field.Name {
			if unicode.IsUpper(c) && i != 0 {
				name += "-"
			}
			name += string(unicode.ToLower(c))
		}
	}
	result := &fieldDesc{
		name:  name,
		usage: usage,
	}
	for _, tag := range infoParts[1:] {
		switch tag {
		case "arg":
			result.positional = true
		case "required":
			result.required = true
		default:
			panic(fmt.Sprintf("unrecognized CLI tag: %s", tag))
		}
	}
	return result
}

type BinaryValue []byte

func (bv *BinaryValue) String() string {
	if bv == nil || len(*bv) == 0 {
		return ""
	}
	return "\\x" + strings.ToUpper(hex.EncodeToString(*bv))
}

func (bv *BinaryValue) Get() interface{} {
	return []byte(*bv)
}

func (bv *BinaryValue) Set(value string) error {
	if value == "" || value[0] != '\\' {
		*bv = []byte(value)
		return nil
	} else if value[1] != 'x' {
		return fmt.Errorf("invalid byte string literal: expected \\x")
	}
	decoded, err := hex.DecodeString(value[2:])
	if err != nil {
		return err
	}
	*bv = decoded
	return nil
}

type StringSliceValue []string

func (ssv *StringSliceValue) String() string {
	if ssv == nil || len(*ssv) == 0 {
		return ""
	}
	return strings.Join(*ssv, ", ")
}

func (ssv *StringSliceValue) Get() interface{} {
	return []string(*ssv)
}

func (ssv *StringSliceValue) Set(value string) error {
	*ssv = append(*ssv, value)
	return nil
}

type flags struct {
	name         string
	flags        *flag.FlagSet
	args         *flag.FlagSet
	reqFlags     map[string]bool
	argsOrder    []string
	firstOptArg  int
	restArgIndex int
}

func newFlags(name string) *flags {
	result := &flags{
		name:         name,
		flags:        flag.NewFlagSet(name, flag.ContinueOnError),
		args:         flag.NewFlagSet(name, flag.PanicOnError),
		reqFlags:     nil,
		argsOrder:    []string{},
		firstOptArg:  0,
		restArgIndex: -1,
	}
	result.flags.Usage = result.Usage
	result.args.Usage = result.Usage
	return result
}

func (f *flags) VisitAll(cb func(*flag.Flag, bool, bool, bool)) {
	f.flags.VisitAll(func(fl *flag.Flag) { cb(fl, false, f.reqFlags[fl.Name], false) })
	for i, name := range f.argsOrder {
		cb(f.args.Lookup(name), true, (i < f.firstOptArg), (i == f.restArgIndex))
	}
}

func (f *flags) Usage() {
	buf := &bytes.Buffer{}
	name := f.name
	if name == "" {
		name = "..."
	}
	fmt.Fprintf(buf, "USAGE: %s", name)

	f.VisitAll(func(fl *flag.Flag, positional, required, rest bool) {
		buf.WriteByte(' ')
		if !required {
			buf.WriteByte('[')
		}
		if positional {
			buf.WriteString(fl.Name)
			if rest {
				buf.WriteString(" ...")
			}
		} else {
			valName, _ := flag.UnquoteUsage(fl)
			buf.WriteByte('-')
			buf.WriteString(fl.Name)
			if valName != "" {
				buf.WriteByte(' ')
				buf.WriteString(valName)
			}
		}
		if !required {
			buf.WriteByte(']')
		}
	})

	fmt.Fprintln(f.flags.Output(), buf.String())
}

func (f *flags) SetOutput(w io.Writer) {
	f.flags.SetOutput(w)
	f.args.SetOutput(w)
}

func (f *flags) failf(format string, values ...interface{}) error {
	err := fmt.Errorf(format, values...)
	fmt.Fprintln(f.flags.Output(), err)
	f.Usage()
	return err
}

func (f *flags) Parse(argv []string) error {
	if err := f.flags.Parse(argv); err != nil {
		return err
	}
	argv = f.flags.Args()

	argIndex := 0
	for _, arg := range argv {
		if argIndex >= len(f.argsOrder) {
			return f.failf("too many positional arguments")
		}
		argName := f.argsOrder[argIndex]
		if err := f.args.Set(argName, arg); err != nil {
			return f.failf("invalid value %q or argument %s: %v",
				arg, argName, err)
		}
		if argIndex != f.restArgIndex {
			argIndex++
		}
	}

	seenFlags := map[string]bool{}
	f.flags.Visit(func(fl *flag.Flag) { seenFlags[fl.Name] = true })
	for name, required := range f.reqFlags {
		if required && !seenFlags[name] {
			return f.failf("missing value for required flag -%s",
				name)
		}
	}
	if argIndex < f.firstOptArg {
		return f.failf("missing value for required argument %s",
			f.argsOrder[argIndex])
	}

	return nil
}

func defaultValueString(f *flag.Flag) (string, bool) {
	zv := reflect.New(reflect.TypeOf(f.Value).Elem())
	zvs := zv.Interface().(flag.Value).String()
	if f.DefValue == zvs {
		return "", false
	}

	if zv.Kind() == reflect.String {
		return fmt.Sprintf("%q", f.DefValue), true
	} else {
		return f.DefValue, true
	}
}

// Yes, that name is stupid. No, I did not invent it.
func (f *flags) PrintDefaults() {
	f.flags.PrintDefaults()

	buf := &bytes.Buffer{}
	for i, name := range f.argsOrder {
		argFlag := f.args.Lookup(name)
		valName, usage := flag.UnquoteUsage(argFlag)
		fmt.Fprintf(buf, "  %s", name)
		if valName != "" && valName != name {
			fmt.Fprintf(buf, ": %s", valName)
		}
		fmt.Fprintf(buf, "\n    \t%s",
			strings.ReplaceAll(usage, "\n", "\n    \t"))
		if i < f.firstOptArg {
			// Required arguments' defaults are of little use.
		} else if text, show := defaultValueString(argFlag); show {
			fmt.Fprintf(buf, " (default %s)", text)
		}
		fmt.Fprintln(buf)
	}
	fmt.Fprint(f.args.Output(), buf.String())
}

type CLIEnv interface {
	Console
	Parent() CLIEnv
	Commands() CommandSet
	Stop()
}

type CommandParams interface {
	Run(env CLIEnv)
}

type Command struct {
	Name     string
	Defaults CommandParams
}

func (c *Command) flags() (*flags, CommandParams) {
	flagGetterType := reflect.TypeOf((*flag.Getter)(nil)).Elem()
	unknownType := func(field reflect.StructField) {
		panic("unsupported command parameter type: " + field.Type.String())
	}

	result := newFlags(c.Name)

	defaults := reflect.ValueOf(c.Defaults)
	tp := defaults.Type()
	if tp.Kind() != reflect.Pointer {
		panic("command parameters must be pointer to struct, got " +
			tp.Kind().String())
	}
	defaults = defaults.Elem()
	tp = defaults.Type()
	if tp.Kind() != reflect.Struct {
		panic("command parameters must be pointer to struct, got pointer to " +
			tp.Kind().String())
	}

	params := reflect.New(tp).Elem()
	params.Set(defaults)

	positionalRequired := true
	for _, field := range reflect.VisibleFields(tp) {
		if field.Anonymous || field.PkgPath != "" {
			continue
		}

		fd := newFieldDesc(field)
		if fd == nil {
			continue
		}

		ov := defaults.FieldByIndex(field.Index).Interface()
		vp := params.FieldByIndex(field.Index).Addr().Interface()
		flags := result.flags
		if fd.positional {
			if result.restArgIndex != -1 {
				panic("cannot have positional argument after rest argument")
			}
			result.argsOrder = append(result.argsOrder, fd.name)
			if fd.required {
				if !positionalRequired {
					panic("required positional argument cannot follow optional one")
				}
				result.firstOptArg = len(result.argsOrder)
			} else {
				positionalRequired = false
			}
			flags = result.args
		} else if fd.required {
			if result.reqFlags == nil {
				result.reqFlags = map[string]bool{fd.name: true}
			} else {
				result.reqFlags[fd.name] = true
			}
		}

		isRest := false
		switch field.Type.Kind() {
		case reflect.Bool:
			flags.BoolVar(vp.(*bool), fd.name, ov.(bool), fd.usage)
		case reflect.Int:
			flags.IntVar(vp.(*int), fd.name, ov.(int), fd.usage)
		case reflect.Int64:
			flags.Int64Var(vp.(*int64), fd.name, ov.(int64), fd.usage)
		case reflect.Uint:
			flags.UintVar(vp.(*uint), fd.name, ov.(uint), fd.usage)
		case reflect.Uint64:
			flags.Uint64Var(vp.(*uint64), fd.name, ov.(uint64), fd.usage)
		case reflect.Float64:
			flags.Float64Var(vp.(*float64), fd.name, ov.(float64), fd.usage)
		case reflect.String:
			flags.StringVar(vp.(*string), fd.name, ov.(string), fd.usage)
		case reflect.Slice:
			switch field.Type.Elem().Kind() {
			case reflect.Uint8:
				flags.Var((*BinaryValue)(vp.(*[]byte)), fd.name, fd.usage)
			case reflect.String:
				flags.Var((*StringSliceValue)(vp.(*[]string)), fd.name, fd.usage)
				isRest = true
			default:
				unknownType(field)
			}
		case reflect.Struct:
			if reflect.PointerTo(field.Type).Implements(flagGetterType) {
				flags.Var(vp.(flag.Getter), fd.name, fd.usage)
			} else {
				unknownType(field)
			}
		default:
			unknownType(field)
		}

		if fd.positional && isRest {
			result.restArgIndex = len(result.argsOrder) - 1
		}
	}

	return result, params.Addr().Interface().(CommandParams)
}

func (c *Command) Parse(con Console, argv []string) CommandParams {
	flags, params := c.flags()
	flags.SetOutput(con)
	err := flags.Parse(argv)
	if err == flag.ErrHelp {
		// flags has already written the usage, append the help
		flags.PrintDefaults()
		return nil
	} else if err != nil {
		// flags has already written an error message
		return nil
	}
	return params
}

func (c *Command) Run(env CLIEnv, argv []string) {
	params := c.Parse(env, argv)
	if params == nil {
		return
	}
	params.Run(env)
}

func (c *Command) Help(con Console) {
	flags, _ := c.flags()
	flags.SetOutput(con)
	flags.Usage()
	flags.PrintDefaults()
}

type CommandSet map[string]*Command

func (cs CommandSet) Add(cmd *Command) { cs[cmd.Name] = cmd }

func (cs CommandSet) GetCommand(con Console, name string) *Command {
	cmd, ok := cs[name]
	if !ok {
		con.Println("unknown command " + name)
		return nil
	}
	return cmd
}

func (cs CommandSet) Run(env CLIEnv, argv []string) {
	if len(argv) == 0 {
		return
	}
	cmd := cs.GetCommand(env, argv[0])
	if cmd == nil {
		return
	}
	cmd.Run(env, argv[1:])
}

type HelpCmd struct {
	Command string `usage:"command to get help of" cli:",arg"`
}

func (c *HelpCmd) Run(env CLIEnv) {
	if c.Command == "" {
		allNames := []string{}
		for name, _ := range env.Commands() {
			allNames = append(allNames, name)
		}
		sort.Strings(allNames)
		msg := []byte("Known commands: ")
		for i, name := range allNames {
			if i != 0 {
				msg = append(msg, ", "...)
			}
			msg = append(msg, name...)
		}
		env.Println(string(msg))
	} else {
		cmd := env.Commands().GetCommand(env, c.Command)
		if cmd == nil {
			return
		}
		cmd.Help(env)
	}
}

type QuitCmd struct {}

func (c *QuitCmd) Run(env CLIEnv) {
	env.Stop()
}

func parseWord(word string) string {
	if word[0] != '"' {
		return word
	}

	result := []byte{}
	for offset := 1; offset < len(word)-1; {
		shift := strings.IndexByte(word[offset:], '\\')
		if shift == -1 {
			result = append(result, word[offset:len(word)-1]...)
			break
		}
		result = append(result, word[offset:offset+shift]...)
		offset += shift

		nextOffset := offset + 2
		if word[offset+1] == '\\' || word[offset+1] == '"' {
			offset++
		}
		result = append(result, word[offset:nextOffset]...)
		offset = nextOffset
	}
	return string(result)
}

func SplitLine(line string) ([]string, error) {
	offset := 0
	fail := func(msg string) ([]string, error) {
		return nil, fmt.Errorf("offset %d: %s", offset, msg)
	}
	result := []string{}
	for first := true; offset != len(line); first = false {
		if !first {
			m := spaceRe.FindStringIndex(line[offset:])
			if m == nil || m[0] != 0 {
				return fail("unexpected characters")
			}
			offset += m[1]
		}
		m := wordRe.FindStringIndex(line[offset:])
		if m == nil {
			return fail("syntax error")
		} else if m[0] != 0 {
			return fail("unexpected characters")
		}
		result = append(result, parseWord(line[offset:offset+m[1]]))
		offset += m[1]
	}
	return result, nil
}

type CLI struct {
	Console
	Prompt   string
	Argv     []string `usage:"single command to run" cli:",arg"`
	parent   CLIEnv
	commands CommandSet
	stop     bool
}

func NewCLI(prompt string, standardCommands bool, commands... *Command) *CLI {
	result := &CLI{Prompt: prompt, commands: CommandSet{}}
	if standardCommands {
		result.AddStandardCommands()
	}
	for _, cmd := range commands {
		result.commands.Add(cmd)
	}
	return result
}

func (c *CLI) Parent() CLIEnv {
	return c.parent
}

func (c *CLI) Commands() CommandSet {
	return c.commands
}

func (c *CLI) Stop() {
	c.stop = true
}

func (c *CLI) AddStandardCommands() {
	c.AddNewCommand("help", &HelpCmd{})
	c.AddNewCommand("quit", &QuitCmd{})
}

func (c *CLI) AddCommand(cmd *Command) {
	c.commands.Add(cmd)
}

func (c *CLI) AddNewCommand(name string, defaults CommandParams) {
	c.AddCommand(&Command{Name: name, Defaults: defaults})
}

func (c *CLI) runOne(argv []string) {
	c.commands.Run(c, argv)
}

func (c *CLI) runLoop() {
	for !c.stop {
		line := c.ReadLine(c.Prompt)
		if line == nil {
			break
		}
		argv, err := SplitLine(*line)
		if err != nil {
			c.Println(err.Error())
			continue
		}
		c.runOne(argv)
	}
}

func (c *CLI) Run(parent CLIEnv) {
	if c.parent != nil {
		panic("trying to run CLI already bound to parent")
	}

	c.Console = parent
	c.parent = parent

	if len(c.Argv) == 0 {
		c.runLoop()
	} else {
		c.runOne(c.Argv)
	}
}

type launcher struct {
	Console
}

func (l launcher) Parent() CLIEnv {
	return nil
}

func (l launcher) Commands() CommandSet {
	return nil
}

func (l launcher) Stop() {}

func Launch(cli *CLI, con Console, argv []string) {
	cli.Argv = argv
	cli.Run(launcher{con})
}

func LaunchAsCommand(cli *CLI, con Console, name string, argv []string) {
	(&Command{name, cli}).Run(launcher{con}, argv)
}

func NormalizeProgName(argv0 string) string {
	return filepath.Base(argv0)
}

func LaunchOS(cli *CLI) {
	con := NewDefaultConsole()
	defer con.Close()

	LaunchAsCommand(cli, con, NormalizeProgName(os.Args[0]), os.Args[1:])
}
