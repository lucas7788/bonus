DROP TABLE IF EXISTS `bonus_transaction_info`;
create table bonus_transaction_info
(
 Id int(11) NOT NULL AUTO_INCREMENT COMMENT '主键',
 TokenType varchar(20) not null COMMENT 'token type',
 EventType varchar(100) not null COMMENT 'event type',
 ContractAddress varchar(100) not null COMMENT 'contract address',
 Address  varchar(100) not null COMMENT 'address',
 Amount varchar(100) not null COMMENT 'amount',
 TxHash varchar(100) not null DEFAULT "" COMMENT 'txHash',
 TxTime bigint(20) NOT NULL DEFAULT 0 COMMENT 'tx time',
 TxHex varchar(5000) not null default "",
 ErrorDetail varchar(1000) not null default "",
 TxResult tinyint(1) NOT NULL DEFAULT 0 COMMENT '交易状态0:build tx failed,1:not send, 2:send failed, 3:send success,4:tx failed,5:tx success',
 PRIMARY KEY (Id)
);
