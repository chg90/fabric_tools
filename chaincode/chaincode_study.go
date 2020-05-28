package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// 必须导入fabric的shim包及peer包，并实现其Init及Invoke方法
// 这两个方法要以一个结构体(可以理解成对象)为载体，可以是空的结构体
type ChaincodeStudy struct {
}

// main方法也是合约必备的方法，是启动整个合约的入口
func main() {
	if err := shim.Start(new(ChaincodeStudy)); err != nil {
		fmt.Println("chaincode start error ", err)
	}

}

// Init 方法会在执行peer chaincode instantiate命令的时候调用
func (t *ChaincodeStudy) Init(stub shim.ChaincodeStubInterface) pb.Response {
	/*
		stub.GetFunctionAndParameters()方法可以获得传入的方法名及参数
		例如使用下列初始化命令:
			peer chaincode instantiate -o  orderer.caohuan.com:7050
			-C caohuanchannel -n caohuanTestcc -v 1.0 -c '{"Args":["init","a","100","b","500"]}'
			-P " OR	('Org1MSP.member','Org2MSP.member')"
		funcName打印出来就是"init"  args就是["a","100","b","500"]
		所以此处命令行的init可以换成任意方法名，只要在合约的Init()方法中指定该方法即可
	*/
	funcName, args := stub.GetFunctionAndParameters()
	fmt.Printf("funcname: %v ,args: %v", funcName, args)

	// return shim.Success代表执行成功
	//如果有错误，可以return shim.Error("something is wrong")
	return shim.Success(nil)
}

/*
	peer chaincode invoke 和 peer chaincode query 都是进入到Invoke方法
	两者区别在于：
		invoke要适应合约的背书规则，去相应的背书节点获取签名从而把交易发送给order节点进行打包
		query不会提交交易，不会发往背书节点，更不会上链
	所以使用query进行查询，而调用交易一定要使用invoke命令
*/
func (t *ChaincodeStudy) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// 同样的，通过stub.GetFunctionAndParameters()获取函数名及参数
	// 这样是为了避免把所有的实现都写在同一个方法内，根据不同的funcName再去调用不同的函数
	funcName, args := stub.GetFunctionAndParameters()
	if funcName == "invoke" {
		return t.invokeCC(stub, args)
	} else if funcName == "query" {
		return t.queryCC(stub, args)
	}

	return shim.Success(nil)
}

func (t *ChaincodeStudy) invokeCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	return shim.Success(nil)
}

func (t *ChaincodeStudy) queryCC(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// return的是一个[]byte
	result := []byte("value")
	return shim.Success(result)
}

// 总结一下合约内常用的API
func (t *ChaincodeStudy) ApiList(stub shim.ChaincodeStubInterface) {
	// 获取调用方法及参数
	funcName, args := stub.GetFunctionAndParameters()
	fmt.Println(funcName, args)

	// 存值
	err := stub.PutState("key", []byte("value"))
	fmt.Println(err)

	// 取值
	res, err := stub.GetState("key")
	fmt.Println(res, err)

	// 删除
	err = stub.DelState("key")

	/*
		用于同一channel中对于不同的org的权限控制，用的较少。
		需要在instantiate chaincode的时候就指定好策略：
			{
				"name":"collecionPrivate",
				"policy":"OR('Org1MSP.member')",
				"requiredPeerCount":0,
				"maxPeerCount":3,
				"blockTOLive":3,
				"menberOnlyRead":true
			}

		peer chaincode  instantiate --collections-config 实例化的时候指定文件路径
	*/
	err = stub.PutPrivateData("collecionPrivate", "key", []byte("value"))
	res, err = stub.GetPrivateData("collecionPrivate", "key")
	err = stub.DelPrivateData("collecionPrivate", "key")

	// 取值时根据范围取，返回的是一个迭代，只有key有序的情况下才有用
	iterator, err := stub.GetStateByRange("startKey", "endKey")
	if iterator.HasNext() {
		fmt.Println(iterator.Next())
	}

	// couchDB的复查询功能，可以不仅根据key来查，还可以根据value来查，没有索引的情况下比较慢
	// 例如查询所有 age = 20 的人
	iterator, err = stub.GetQueryResult(fmt.Sprintf("{\"selector\":{\"age\":\"%v\"}}", 20))

	// 根据取值范围查询，并分页,bookmark类似于mysql中的offset
	iterator, bookmark, err := stub.GetStateByRangeWithPagination("startKey", "endKey", 10, "")
	fmt.Println(bookmark)

	// 根据复查询分页
	iterator, bookmark, err = stub.GetQueryResultWithPagination("sql", 10, "")

	//获取某个key所有的历史记录
	hisIterator, err := stub.GetHistoryForKey("key")
	if hisIterator.HasNext() {
		fmt.Println(hisIterator.Next())
	}

	// 设置event，可以发布/订阅消息队列
	var as []byte
	for _, a := range args {
		as = append(as, []byte(a)...)
	}
	stub.SetEvent("InitEvent", as)

	//调用其他合约
	stub.InvokeChaincode("合约名", [][]byte{[]byte("invoke"), []byte("transfer"), []byte("a")}, "通道")

	//设置日志级别
	//将"debug"字符串转换成LoggingLevel类型
	//"CRITICAL"      0
	//"ERROR"         1
	//"WARNING"       2
	//"NOTICE"        3
	//"INFO"          4
	//"DEBUG"         5
	logLevel, _ := shim.LogLevel("DEBUG")
	shim.SetLoggingLevel(logLevel)

}
