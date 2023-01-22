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

    miner_ids = []
    N = 0

    @staticmethod
    def parse_block(str_block) -> "Block":
        if len(str_block) == 0:
            return Block(0, 0, 0, "", "")

        if not ("Block:" in str_block and "Timestamp:" in str_block):
            return Block(0, 0, 0, "", "")

        block = Block(_id=0, timestamp=0, magic=0, hash_prev="", _hash="")

        lines = [line.strip() for line in str_block.splitlines() if len(line.strip()) > 0]

        if len(lines) != 11:
            raise BlockParseException("Every block should contain 11 lines of data")

        if lines[0] != "Block:":
            raise BlockParseException("First line of every block should be \"Block:\"")

        if not lines[1].startswith("Created by"):
            raise BlockParseException("Second line of every block should start with \"Created by\"")

        block.miner_ids.append(lines[1])

        if not lines[2].startswith("Id:"):
            raise BlockParseException("Third line of every block should start with \"Id:\"")

        _id = lines[2].split(":")[1].strip().replace("-", "")
        if not _id.isdigit():
            raise BlockParseException("Id should be a number")

        block._id = int(_id)

        if not lines[3].startswith("Timestamp:"):
            raise BlockParseException("4-th line of every block should start with \"Timestamp:\"")

        timestamp = lines[3].split(":")[1].strip().replace("-", "")
        if not timestamp.isdigit():
            raise BlockParseException("Timestamp should be a number")

        block.timestamp = int(timestamp)

        if not lines[4].startswith("Magic number:"):
            raise BlockParseException("5-th line of every block should start with \"Magic number:\"")

        magic = lines[4].split(":")[1].strip().replace("-", "")
        if not magic.isdigit():
            raise BlockParseException("Magic number should be a number")

        block.magic = int(magic)

        if lines[5] != "Hash of the previous block:":
            raise BlockParseException("6-th line of every block should be \"Hash of the previous block:\"")

        if lines[7] != "Hash of the block:":
            raise BlockParseException("8-th line of every block should be \"Hash of the block:\"")

        if not lines[10].upper().startswith("N "):
            raise BlockParseException("11-th line of every block should be state what happened to N")

        prev_hash = lines[6].strip()
        _hash = lines[8].strip()

        if not (len(prev_hash) == 64 or prev_hash == "0") or not len(_hash) == 64:
            raise BlockParseException("Hash length should be equal to 64 except \"0\"")

        if _hash == prev_hash:
            raise BlockParseException("The current hash and the previous hash in a block should be different.")

        if not _hash.startswith("0" * block.N):
            raise BlockParseException(f"N is {block.N} but hash, {_hash} , did not start with the correct number of "
                                      f"zeros.")

        block.hash_prev = prev_hash
        block._hash = _hash

        if "increase" in lines[10].lower():
            block.N += 1
        elif "decrease" in lines[10].lower():
            block.N -= 1
            if block.N < 0:
                raise BlockParseException("N was decreased below zero!")

        elif "same" not in lines[10].lower():
            raise BlockParseException("11-th line of every block must state "
                                      "N increased, decreased, or stayed the same.")

        return block

    @staticmethod
    def parse_blocks(output) -> list["Block"]:
        Block.N = 0

        str_blocks = output.split("\n\n")

        blocks = []

        for str_block in str_blocks:
            block = Block.parse_block(str_block)
            if block is not None:
                blocks.append(block)

        first_miner = Block.miner_ids[0]
        Block.miner_ids = [miner for miner in Block.miner_ids if miner != first_miner]
        if len(Block.miner_ids) == 0:
            raise BlockParseException("All blocks are mined by a single miner!")

        return blocks


class Clue:
    def __init__(self, n):
        self.zeros = "0" * n


class BlockchainTest(StageTest):
    previous_outputs = []

    def generate(self) -> list[TestCase]:
        return [TestCase(), TestCase()]

    def check(self, reply: str, clue: Clue) -> CheckResult:
        if reply in self.previous_outputs:
            return CheckResult(False, "You already printed this text in the previous tests")

        self.previous_outputs.append(reply)

        try:
            blocks = Block.parse_blocks(reply)
        except BlockParseException as ex:
            return CheckResult.wrong(str(ex))
        except Exception as ex:
            return CheckResult.wrong("Something went wrong while parsing the output. " + str(ex))

        if len(blocks) != 5:
            return CheckResult(False, f"You should output 5 blocks, found {len(blocks)}")

        for i in range(1, len(blocks)):
            curr = blocks[i - 1]
            _next = blocks[i]

            if curr._id + 1 != _next._id:
                return CheckResult(False, "Id's of blocks should increase by 1")

            if _next.timestamp < curr.timestamp:
                return CheckResult(False, "Timestamp`s of blocks should increase")

            if _next.hash_prev != curr._hash:
                return CheckResult(False, "Two hashes aren't equal, but should")

        return CheckResult.correct()
