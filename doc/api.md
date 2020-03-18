
# 接口文档
1. 上传excel   POST

http://127.0.0.1:8080/api/v1/uploadexecl   

参数
```json
{
 "id": 1,
    "jsonrpc": "2.0",
    "method": "uploadexecl",
    "params": {
     "billList": [
     {
       "address": "0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648",
       "amount": "2"
     },
     {
       "address":"0xeE25bD734e4cfacdb98aE1454A0Cd4d259e69513",
       "amount": "1"
     },
     {
     	"address":"0x35401776e5B2d2b331363b5aAFCEa2eAe8B04546",
     	"amount":"1"
     },
     {
     	"address":"0xbcc1223369b23771a94cccD1cf2E9fbD8bEC0251",
     	"amount":"1"
     },
     {
     	"address":"0x95FB49AE2DEC0D2a37b27033742fd99915faF6A1",
     	"amount":"1"
     }
   ],
   "tokenType": "ERC20",
   "contractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
   "eventType": "ssserc20",
   "netType":"testnet"
 }
}
```

响应
```json
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": {
        "BillList": [
            {
                "Id": 1,
                "Address": "0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648",
                "Amount": "2"
            },
            {
                "Id": 2,
                "Address": "0xeE25bD734e4cfacdb98aE1454A0Cd4d259e69513",
                "Amount": "1"
            },
            {
                "Id": 3,
                "Address": "0x35401776e5B2d2b331363b5aAFCEa2eAe8B04546",
                "Amount": "1"
            },
            {
                "Id": 4,
                "Address": "0xbcc1223369b23771a94cccD1cf2E9fbD8bEC0251",
                "Amount": "1"
            },
            {
                "Id": 5,
                "Address": "0x95FB49AE2DEC0D2a37b27033742fd99915faF6A1",
                "Amount": "1"
            }
        ],
        "TokenType": "ERC20",
        "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
        "EventType": "ssserc20",
        "Admin": "0x6912A9fdB690ecfe10ecaa4B71141B019075807B",
        "EstimateFee": "0.0042",
        "Sum": "6",
        "AdminBalance": {
            "ERC20": "0",
            "ETH": "0"
        },
        "NetType": "testnet",
        "Total": 5
    },
    "Version": "1.0.0"
}
```


2. 查询已经上传的所有的eventtype   GET

请求
```
http://127.0.0.1:8080/api/v1/getallevtty
```
响应
```json
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": [
        "ssserc20"
    ],
    "Version": "1.0.0"
}
```

3. 根据eventtype、nettype、pageNum、pageSize查询历史上传的excel数据

`/api/v1/getexcelparam/<evtty>/<netty>/<pagenum>/<pagesize>`

pagenum 和pagesize 同时为零  表示查询所有数据

请求
```
http://127.0.0.1:8080/api/v1/getexcelparam/ssserc20/testnet/1/10
```
响应
```
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": {
        "BillList": [
            {
                "Id": 5,
                "Address": "0x95FB49AE2DEC0D2a37b27033742fd99915faF6A1",
                "Amount": "1"
            },
            {
                "Id": 4,
                "Address": "0xbcc1223369b23771a94cccD1cf2E9fbD8bEC0251",
                "Amount": "1"
            },
            {
                "Id": 3,
                "Address": "0x35401776e5B2d2b331363b5aAFCEa2eAe8B04546",
                "Amount": "1"
            },
            {
                "Id": 2,
                "Address": "0xeE25bD734e4cfacdb98aE1454A0Cd4d259e69513",
                "Amount": "1"
            },
            {
                "Id": 1,
                "Address": "0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648",
                "Amount": "2"
            }
        ],
        "TokenType": "ERC20",
        "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
        "EventType": "ssserc20",
        "Admin": "0x6912A9fdB690ecfe10ecaa4B71141B019075807B",
        "EstimateFee": "0.0042",
        "Sum": "6",
        "AdminBalance": {
            "ERC20": "0",
            "ETH": "0"
        },
        "NetType": "",
        "Total": 5
    },
    "Version": "1.0.0"
}
```

4. 查询管理员余额  GET

`/api/v1/getadminbalance/<evtty>/<netty>`

请求
```
http://127.0.0.1:8080/api/v1/getadminbalance/ssserc20/testnet
```
响应
```json
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": {
        "ERC20": "0",
        "ETH": "0"
    },
    "Version": "1.0.0"
}

```

5. 查询GasPrice  GET,   eth的单位是Gwei

``

请求
```
http://127.0.0.1:8080/api/v1/getgasprice/ETH
```

响应
```json
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": 40,
    "Version": "1.0.0"
}
```

6. 修改gasprice

`http://127.0.0.1:8080/api/v1/setgasprice`
参数
```json
{
	"id": 1,
    "jsonrpc": "2.0",
    "method": "setgasprice",
    "params": {
    	"gasPrice":50,
    	"tokenType":"ETH"
    }
}
```

响应
```
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": "",
    "Version": "1.0.0"
}
```

7. 查询转账进度 GET

`/api/v1/gettransferprogress/<evtty>/<netty>`

参数`evtty`是事件类型
`netty`是网络类型

请求
```
http://127.0.0.1:8080/api/v1/gettransferprogress/ssserc20/testnet
```
响应

```
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": {
        "failed": 0,
        "notSend": 0,
        "sendFailed": 0,
        "success": 0,
        "total": 5,
        "transfering": 0
    },
    "Version": "1.0.0"
}
```

8. 转帐 POST
`http://127.0.0.1:8080/api/v1/transfer`

参数
```
{
	"id": 1,
    "jsonrpc": "2.0",
    "method": "transfer",
    "params": {
    	"eventType":"ssserc20",
    	"netType":"testnet"
    }
}
```

响应
```
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": "",
    "Version": "1.0.0"
}
```

9 查询交易有关的所有eventType   GET

`/api/v1/gettxInfoevtty/<netty>`

请求
```
http://127.0.0.1:8080/api/v1/gettxInfoevtty/testnet
```
响应
```
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": [
        "ssserc20"
    ],
    "Version": "1.0.0"
}
```

10 查询交易历史数据  GET
`/api/v1/gettxinfo/<evtty>/<netty>/<pagenum>/<pagesize>`

参数
`evtty`事件类型
`netty` 网络类型
`pagenum` 第几页
`pagesize` 每页的大小

pagenum 和pagesize 同时为零  表示查询所有数据

```
http://127.0.0.1:8080/api/v1/gettxinfo/ssserc20/testnet/1/10
```

响应

```json
{
    "Action": "",
    "Desc": "SUCCESS",
    "Error": 1,
    "Result": {
        "TxInfo": [
            {
                "Id": 1,
                "NetType": "",
                "EventType": "ssserc20",
                "TokenType": "ERC20",
                "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
                "Address": "0x4e7946D1Ee8f8703E24C6F3fBf032AD4459c4648",
                "Amount": "2",
                "TxHash": "0x326761320fa2ba6cbf924fed5b5ca477053b68a523d15bbd9e3b47815d0ded1e",
                "TxTime": 1584521412,
                "TxHex": "",
                "TxResult": 5,
                "ErrorDetail": "success"
            },
            {
                "Id": 2,
                "NetType": "",
                "EventType": "ssserc20",
                "TokenType": "ERC20",
                "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
                "Address": "0xeE25bD734e4cfacdb98aE1454A0Cd4d259e69513",
                "Amount": "1",
                "TxHash": "0xc3c301a8bf07f0b006f1b854eca1f7ca4b9c1a3926458f4ea606accfea2e1037",
                "TxTime": 1584521432,
                "TxHex": "",
                "TxResult": 5,
                "ErrorDetail": "success"
            },
            {
                "Id": 3,
                "NetType": "",
                "EventType": "ssserc20",
                "TokenType": "ERC20",
                "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
                "Address": "0x35401776e5B2d2b331363b5aAFCEa2eAe8B04546",
                "Amount": "1",
                "TxHash": "0xfbc97a1085dd90372786171c644333f81d9a0d2d67bf1921b7dbecb665f96686",
                "TxTime": 1584521432,
                "TxHex": "",
                "TxResult": 5,
                "ErrorDetail": "success"
            },
            {
                "Id": 4,
                "NetType": "",
                "EventType": "ssserc20",
                "TokenType": "ERC20",
                "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
                "Address": "0xbcc1223369b23771a94cccD1cf2E9fbD8bEC0251",
                "Amount": "1",
                "TxHash": "0x77f65f9772a19a19e06312eab5a653409ff39675d20e920962c486c83dbe16e4",
                "TxTime": 1584521432,
                "TxHex": "",
                "TxResult": 5,
                "ErrorDetail": "success"
            },
            {
                "Id": 5,
                "NetType": "",
                "EventType": "ssserc20",
                "TokenType": "ERC20",
                "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
                "Address": "0x95FB49AE2DEC0D2a37b27033742fd99915faF6A1",
                "Amount": "1",
                "TxHash": "0xbeec3085698400839081c7d32b514425b46557e446792b20f274c8dc04189106",
                "TxTime": 1584521432,
                "TxHex": "",
                "TxResult": 5,
                "ErrorDetail": "success"
            }
        ],
        "Admin": "0x6912A9fdB690ecfe10ecaa4B71141B019075807B",
        "EstimateFee": "0.0084",
        "Sum": "6",
        "AdminBalance": {
            "ERC20": "8",
            "ETH": "0.0141592"
        },
        "EventType": "ssserc20",
        "TokenType": "ERC20",
        "ContractAddress": "0x7f3630ce620FD15DE7367207dBFfB085a3F7d118",
        "NetType": "testnet"
    },
    "Version": "1.0.0"
}
```