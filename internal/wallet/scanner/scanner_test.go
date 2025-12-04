package scanner

import (
	"fmt"
	"testing"

	"github.com/jinmu/go-blockchain/internal/wallet/rpc"
)

type scanner interface{
	test(string) string
	setName(string);
	getName() string;
}

type testscanner struct {
	name string

}
func (t *testscanner) test(s string) string {
	return s
}

func (t *testscanner) setName(name string) {
	t.name = name
}

func (t *testscanner) getName() string {
	return t.name
}


type testscanner2 struct {
	name string
}

func (t *testscanner2) test(s string) string {
	return s+"2"
}

func (t *testscanner2) setName(name string) {
	t.name = name
}

func (t *testscanner2) getName() string {
	return t.name
}

type ScannerTestNew struct {
	rpcClient rpc.Client
	username string
	password string
	scanner scanner
}

func TestScanner(t *testing.T) {
	t.Log("TestScanner - placeholder test")
	// TODO: 添加实际的测试用例

	// testscanner := &testscanner{}
	// testscanner2 := &testscanner2{}
	// s1 := testscanner.test("1")
	// s2 := testscanner2.test("1")
	// fmt.Println(s1, s2)


	scannerTest := &ScannerTestNew{
		scanner: &testscanner2{},
		rpcClient: &rpc.EVMClient{},
		username: "",
		password: "",
	}
	scannerTest.scanner.setName("1")
	fmt.Println("--------------------------------")
	fmt.Println(scannerTest.scanner.getName() + "--------------------------------")

}


