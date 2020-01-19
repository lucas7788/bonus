pragma solidity ^0.4.25;

contract Erc20 {
    uint256 public totalSupply;

    uint256 public decimals;

    function balanceOf(address _owner) public constant returns (uint balance);

    function transfer(address _to, uint256 _value) public;

    function transferFrom(address _from, address _to, uint256 _value) public;

    function approve(address _spender, uint256 _value) public;

    function allowance(address _owner, address _spender) public constant returns (uint remaining);

    event Transfer(address indexed _from, address indexed _to, uint256 _value);
    event Approval(address indexed _owner, address indexed _spender, uint256 _value);
}