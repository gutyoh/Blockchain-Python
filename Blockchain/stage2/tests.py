from hstest.stage_test import StageTest
from hstest.test_case import CheckResult
from hstest.test_case import TestCase


class BlockParseException(Exception):
    def __init__(self, msg):
        super().__init__(msg)


class Block:
    def __init__(self, _id, timestamp, magic, hash_prev, _hash):
        self._id = _id
        self.timestamp = timestamp
        self.magic = magic
        self.hash_prev = hash_prev
        self._hash = _hash

    @staticmethod
    def parse_block(str_block) -> 'Block':
        if len(str_block) == 0:
            return Block(_id=0, timestamp=0, magic=0, hash_prev='', _hash='')

        if not ("Block:" in str_block and "Timestamp:" in str_block):
            return Block(_id=0, timestamp=0, magic=0, hash_prev='', _hash='')

        block = Block(_id=0, timestamp=0, magic=0, hash_prev='', _hash='')

        lines = [line.strip() for line in str_block.splitlines() if len(line.strip()) > 0]

        if len(lines) != 9:
            raise BlockParseException("Every block should contain 9 lines of data")

        if lines[0] != "Block:":
            raise BlockParseException("First line of every block should be \"Block:\"")

        if not lines[1].startswith("Id:"):
            raise BlockParseException("Second line of every block should start with \"Id:\"")

        _id = lines[1].split(":")[1].strip().replace("-", "")
        is_numeric = _id.isnumeric()

        if not is_numeric:
            raise BlockParseException("Id should be a number")

        block._id = int(_id)

        if not lines[2].startswith("Timestamp:"):
            raise BlockParseException("Third line of every block should start with \"Timestamp:\"")

        timestamp = lines[2].split(":")[1].strip().replace("-", "")
        is_numeric = timestamp.isnumeric()

        if not is_numeric:
            raise BlockParseException("Timestamp should be a number")

        block.timestamp = int(timestamp)

        if not lines[3].startswith("Magic number:"):
            raise BlockParseException("4-th line of every block should start with \"Magic number:\"")

        magic = lines[3].split(":")[1].strip().replace("-", "")
        is_numeric = magic.isnumeric()

        if not is_numeric:
            raise BlockParseException("Timestamp should be a number")

        block.magic = int(magic)

        if lines[4] != "Hash of the previous block:":
            raise BlockParseException("5-th line of every block should be \"Hash of the previous block:\"")

        if lines[6] != "Hash of the block:":
            raise BlockParseException("7-th line of every block should be \"Hash of the block:\"")

        prev_hash = lines[5].strip()
        _hash = lines[7].strip()

        if not (len(prev_hash) == 64 or prev_hash == "0") or not len(_hash) == 64:
            raise BlockParseException("Hash length should be equal to 64 except \"0\"")

        if _hash == prev_hash:
            raise BlockParseException("The current hash and the previous hash in a block should be different.")

        block._hash = _hash
        block.hash_prev = prev_hash

        return block

    @staticmethod
    def parse_blocks(output: str) -> list['Block']:
        str_blocks = output[output.index("Block:"):].split("\n\n")

        blocks = []

        for str_block in str_blocks:
            block = Block.parse_block(str_block.strip())
            if block is not None:
                blocks.append(block)

        return blocks


class Clue:
    def __init__(self, n):
        self.zeros = "0" * n


class BlockchainTest(StageTest):
    previous_outputs = []

    def generate(self) -> list[TestCase]:
        return [
            TestCase(stdin="0", attach=Clue(0)),
            TestCase(stdin="1", attach=Clue(1)),
            TestCase(stdin="2", attach=Clue(2)),
            TestCase(stdin="0", attach=Clue(0)),
            TestCase(stdin="1", attach=Clue(1)),
            TestCase(stdin="2", attach=Clue(2)),
        ]

    def check(self, reply, clue) -> CheckResult:
        if reply in self.previous_outputs:
            return CheckResult.wrong("You already printed this text in the previous tests")

        self.previous_outputs.append(reply)

        try:
            blocks = Block.parse_blocks(reply)
        except BlockParseException as e:
            return CheckResult.wrong(f"Block parsing error: {e}")
        except Exception as e:
            return CheckResult.wrong(f"Unexpected error: {e}")

        if len(blocks) != 5:
            return CheckResult.wrong("You should output 5 blocks, found {}".format(len(blocks)))

        for i in range(1, len(blocks)):
            curr = blocks[i - 1]
            _next = blocks[i]

            if curr._id + 1 != _next._id:
                return CheckResult.wrong("Id`s of blocks should increase by 1")

            if _next.timestamp < curr.timestamp:
                return CheckResult.wrong("Timestamp`s of blocks should increase")

            if _next.hash_prev != curr._hash:
                return CheckResult.wrong("Two hashes aren't equal, but should")

            if not _next._hash.startswith(clue.zeros):
                return CheckResult.wrong("Hash should start with some zeros")

        return CheckResult.correct()
