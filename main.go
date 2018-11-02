package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	combinator "github.com/jiro4989/colc/combinator/v1"
	colcio "github.com/jiro4989/colc/io"
)

var cs = []combinator.Combinator{
	combinator.Combinator{
		Name:      "S",
		ArgsCount: 3,
		Format:    "{0}{2}({1}{2})",
	},
	combinator.Combinator{
		Name:      "K",
		ArgsCount: 2,
		Format:    "{0}",
	},
	combinator.Combinator{
		Name:      "I",
		ArgsCount: 1,
		Format:    "{0}",
	},
}

// options オプション引数
type options struct {
	Version     func() `short:"v" long:"version" description:"バージョン情報"`
	StepCount   int    `short:"s" long:"stepcount" description:"何ステップまで計算するか" default:"-1"`
	OutFile     string `short:"o" long:"outfile" description:"出力ファイルパス"`
	OutFileType string `short:"t" long:"outfiletype" description:"出力ファイルの種類(なし|json)"`
}

// コンビネータ設定
type Config []CombinatorFormat

type CombinatorFormat struct {
	ArgsCount      int    `json:"argsCount"`
	CombinatorName string `json:"combinatorName"`
	Format         string `json:"format"`
}

// エラー出力ログ
var logger = log.New(os.Stderr, "", 0)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func main() {
	opts, args := parseOptions()

	failure := func(err error) {
		panic(err)
	}

	// 引数指定なしの場合は標準入力を処理
	if len(args) < 1 {
		r := os.Stdin
		if err := calcOut(r, opts, out, failure); err != nil {
			panic(err)
		}
		return
	}
	// 引数指定ありの場合はファイル処理
	for _, fn := range args {
		err := colcio.WithOpen(fn, func(r io.Reader) error {
			return calcOut(r, opts, out, failure)
		})
		if err != nil {
			panic(err)
		}
	}
}

// calcOut はCLCodeを計算して、出力する。
// 計算結果を引数の関数に私、失敗時は引数に渡した関数を適用する。
func calcOut(r io.Reader, opts options, success func([]string, options) error, failure func(error)) error {
	ss, err := calcCLCode(r, opts)
	if err != nil {
		failure(err)
	}
	return success(ss, opts)
}

// calcCLCode はCLCodeを計算し、スライスで返す。
func calcCLCode(r io.Reader, opts options) ([]string, error) {
	var res []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		line = strings.Trim(line, " ")
		s := combinator.CalcCLCode(line, cs, opts.StepCount)
		res = append(res, s)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

// out は行配列をオプションに応じて出力する。
// 出力先ファイルが指定されていなければ標準出力する。
// 指定があればファイル出力する。
func out(lines []string, opts options) error {
	if opts.OutFile == "" {
		for _, v := range lines {
			fmt.Println(v)
		}
		return nil
	}

	return colcio.WriteFile(opts.OutFile, lines)
}

// ReadConfig は指定パスのJSON設定ファイルを読み取る
func ReadConfig(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return config, nil
}

// parseOptions はコマンドラインオプションを解析する。
// 解析あとはオプションと、残った引数を返す。
func parseOptions() (options, []string) {
	var opts options
	opts.Version = func() {
		fmt.Println(Version)
		os.Exit(0)
	}

	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(0)
	}

	return opts, args
}
