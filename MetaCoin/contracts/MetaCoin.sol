pragma solidity >=0.4.25 <0.6.0;

import "./ConvertLib.sol";

// This is just a simple example of a coin-like contract.
// It is not standards compatible and cannot be expected to talk to other
// coin/token contracts. If you want to create a standards-compliant
// token, see: https://github.com/ConsenSys/Tokens. Cheers!

contract MetaCoin {
	struct Htlc {
		bytes32 rHash;
		uint amount;
		address htlcSuccessAddress;
		address htlcTimeoutAddress;
	}

	mapping (address => uint) balances;

	Htlc htlc;

	event Transfer(address indexed _from, address indexed _to, uint256 _value);

	constructor() public {
		balances[tx.origin] = 10000;
	}

	function sendCoin(address receiver, uint amount) public returns(bool sufficient) {
		if (balances[msg.sender] < amount) return false;
		balances[msg.sender] -= amount;
		balances[receiver] += amount;
		emit Transfer(msg.sender, receiver, amount);
		return true;
	}

	function getBalanceInEth(address addr) public view returns(uint){
		return ConvertLib.convert(getBalance(addr),2);
	}

	function getBalance(address addr) public view returns(uint) {
		return balances[addr];
	}

	function createHtlc(bytes32 rHash, uint amount, address htlcSuccessAddress) public returns(bool sufficient) {
		if (balances[msg.sender] < amount) return false;
		balances[msg.sender] -= amount;

		htlc.rHash = rHash;
		htlc.amount = amount;
		htlc.htlcSuccessAddress = htlcSuccessAddress;
		htlc.htlcTimeoutAddress = msg.sender;

		return true;
	}

	function htlcSuccess(bytes memory rPreImage) public returns(bool success) {
		bytes32 calcRHash = sha256(rPreImage);
		if (calcRHash != htlc.rHash) {
			return false;
		}

		balances[htlc.htlcSuccessAddress] += htlc.amount;

		deleteHtlc();
		return true;
	}

	function htlcTimeout() public {
		balances[htlc.htlcTimeoutAddress] += htlc.amount;

		deleteHtlc();
	}

	function deleteHtlc() private {
		delete htlc.rHash;
		delete htlc.amount;
		delete htlc.htlcSuccessAddress;
		delete htlc.htlcTimeoutAddress;
	}
}

contract FakeMetaCoin {
	function fakeCreateHtlc(bytes32 rHash, uint amount, address htlcSuccessAddress) public returns(bool sufficient) {
		return true;
	}
}