from hstest.stage_test import StageTest
from hstest.test_case import CheckResult
from hstest.test_case import TestCase


class BlockParseException(Exception):
    def __init__(self, msg):
        super().__init__(msg)


class Block:
    def __init__(self, _id, timestamp, hash_prev, _hash):
        self._id = _id
        self.timestamp = timestamp
        self.hash_prev = hash_prev
        self._hash = _hash

    @staticmethod
    def parse_block(str_block) -> 'Block':
        if len(str_block) == 0:
            return Block(_id=0, timestamp=0, hash_prev='', _hash='')

        block = Block(_id=0, timestamp=0, hash_prev='', _hash='')

        lines = [line.strip() for line in str_block.splitlines() if len(line.strip()) > 0]

        if len(lines) != 7:
            raise BlockParseException("Every block should contain 7 lines of data")

        if lines[0] != "Block:":
            raise BlockParseException("First line of every block should be \"Block:\"")

        if not lines[1].startswith("Id:"):
            raise BlockParseException("Second line of every block should start with \"Id:\"")

        _id = lines[1].split(":")[1].strip()
        is_numeric = _id.isnumeric()

        if not is_numeric:
            raise BlockParseException("Id should be a number")

        block._id = int(_id)

        if not lines[2].startswith("Timestamp:"):
            raise BlockParseException("Third line of every block should start with \"Timestamp:\"")

        timestamp = lines[2].split(":")[1].strip()
        is_numeric = timestamp.isnumeric()

        if not is_numeric:
            raise BlockParseException("Timestamp should be a number")

        block.timestamp = int(timestamp)

        if lines[3] != "Hash of the previous block:":
            raise BlockParseException("4-th line of every block should be \"Hash of the previous block:\"")

        if lines[5] != "Hash of the block:":
            raise BlockParseException("6-th line of every block should be \"Hash of the block:\"")

        prev_hash = lines[4].strip()
        _hash = lines[6].strip()

        if not (len(prev_hash) == 64 or prev_hash == "0") or not len(_hash) == 64:
            raise BlockParseException("Hash length should be equal to 64 except \"0\"")

        if _hash == prev_hash:
            raise BlockParseException("The current hash and the previous hash in a block should be different.")

        block.hash_prev = prev_hash
        block._hash = _hash

        return block

    @staticmethod
    def parse_blocks(output) -> list['Block']:
        str_blocks = output.split("\n\n")

        blocks = []

        for str_block in str_blocks:
            block = Block.parse_block(str_block.strip())
            if block is not None:
                blocks.append(block)

        return blocks


class BlockchainTest(StageTest):
    previous_outputs = []

    def generate(self) -> list['TestCase']:
        return [TestCase(), TestCase()]

    def check(self, reply, clue) -> 'CheckResult':
        if reply in self.previous_outputs:
            return CheckResult(False, "You already printed this text in the previous tests")

        self.previous_outputs.append(reply)

        blocks = Block.parse_blocks(reply)

        if len(blocks) != 5:
            return CheckResult(False, f"You should output 5 blocks, found {len(blocks)}")

        first = blocks[0]

        if first.hash_prev != "0":
            return CheckResult(False, "Previous hash of the first block should be \"0\"")

        for i in range(1, len(blocks)):
            curr = blocks[i - 1]
            _next = blocks[i]

            if curr._id + 1 != _next._id:
                return CheckResult(False, "Id`s of blocks should increase by 1")

            if _next.timestamp < curr.timestamp:
                return CheckResult(False, "Timestamp`s of blocks should increase")

            if _next.hash_prev != curr._hash:
                return CheckResult(False, "Two hashes aren't equal, but should")

        return CheckResult.correct()
