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
    data_strs = []
    N = 0

    @staticmethod
    def parse_block(str_block) -> "Block":
        if len(str_block) == 0:
            return Block(0, 0, 0, "", "")

        if not ("Block:" in str_block and "Timestamp:" in str_block):
            return Block(0, 0, 0, "", "")

        block = Block(_id=0, timestamp=0, magic=0, hash_prev="", _hash="")

        lines = [line.strip() for line in str_block.splitlines() if len(line.strip()) > 0]

        if len(lines) < 12:
            if block._id > 5:
                raise BlockParseException("The program should ONLY create and print 5 blocks!")
            else:
                raise BlockParseException("Every block should contain at least 12 lines of data")

        if not lines[0].startswith("Block"):
            raise BlockParseException("First line of every block should start with \"Block\"")

        if not lines[1].startswith("Created by"):
            raise BlockParseException("Second line of every block should start with \"Created by\"")

        block.miner_ids.append(lines[1])

        if not lines[2].startswith("Id:"):
            raise BlockParseException("Third line of every block should start with \"Id:\"")

        _id = lines[2].split(":")[1].strip().replace("-", "")
        is_numeric = _id.isnumeric()

        if not is_numeric:
            raise BlockParseException("Id should be a number")

        block._id = int(_id)

        if not lines[3].startswith("Timestamp:"):
            raise BlockParseException("4-th line of every block should start with \"Timestamp:\"")

        timestamp = lines[3].split(":")[1].strip().replace("-", "")
        is_numeric = timestamp.isnumeric()

        if not is_numeric:
            raise BlockParseException("Timestamp should be a number")

        block.timestamp = int(timestamp)

        if not lines[4].startswith("Magic number:"):
            raise BlockParseException("5-th line of every block should start with \"Magic number:\"")

        magic = lines[4].split(":")[1].strip().replace("-", "")
        is_numeric = magic.isnumeric()

        if not is_numeric:
            raise BlockParseException("Magic number should be a number")

        block.magic = int(magic)

        if not lines[5].startswith("Hash of the previous block:"):
            raise BlockParseException("6-th line of every block should start with \"Hash of the previous block:\"")

        if not lines[7].startswith("Hash of the block:"):
            raise BlockParseException("8-th line of every block should start with \"Hash of the block:\"")

        prev_hash = lines[6].strip()
        _hash = lines[8].strip()

        if not (len(prev_hash) == 64 or prev_hash == "0") or not len(_hash) == 64:
            raise BlockParseException("Hash length should be equal to 64 except \"0\"")

        if _hash == prev_hash:
            raise BlockParseException("The current hash and the previous hash in a block should be different.")

        if not _hash.startswith("0" * Block.N):
            raise BlockParseException(
                f"N is {Block.N} but hash, {_hash} , did not start with the correct number of zeros.")

        block._hash = _hash
        block.hash_prev = prev_hash

        #  Check the `Block data` of the first/genesis block:
        if block._id == 1:
            if not lines[9].startswith("Block data:"):
                raise BlockParseException("10-th line of the first block " +
                                          "should start with \"Block data:\"")

            if "no messages" not in lines[9]:
                raise BlockParseException("10-th line of the first block " +
                                          "should contain \"no messages\"")

            if not ("block" in lines[10].lower() or "generating" in lines[10].lower()):
                raise BlockParseException("11-th line of the first block " +
                                          "should say how long the block was generating for! "
                                          "(Use the example's format)")

            if not lines[11].upper().startswith("N "):
                raise BlockParseException("12-th line of the first block " +
                                          "should be state what happened to N in the format given.")

            if not lines[12].startswith("Enter message"):
                raise BlockParseException("The last line of every block " +
                                          "should start with \"Enter message\"")

        # Then check the `Block data of the remaining blocks:`
        if 1 < block._id < 5:
            if not lines[9].lower().startswith("block data:"):
                raise BlockParseException("10-th line of every block should start with \"Block data:\"")

            i = 10

            while not lines[i].lower().startswith("message"):
                i += 1

            if not lines[i].lower().startswith("message"):
                raise BlockParseException("After the line with \"Block data:\" " +
                                          "the next line should contain the message "
                                          "ID and start with \"Message ID:\" " +
                                          "followed by a unique ID for the message(s).")

            i += 1

            if not lines[i].lower().startswith("signature"):
                raise BlockParseException("After the line with \"Message ID:\" " +
                                          "the next line should contain the message signature "
                                          "and start with \"Signature:\" " +
                                          "followed by the message signature(s).")

            # Get the signature value after the `:` colon in the string `Signature:`
            regex = ":"
            signature = lines[i].split(regex, 1)[1].strip()

            if signature is None:
                raise BlockParseException("Make sure you write the signature after the `Signature:` string.\n" +
                                          "For example: \"Signature:MEUCIBFU...\"")

            # Check if the signature starts with `ME` and has a length of 96 characters:
            if not signature.startswith("ME") or len(signature) != 96:
                raise BlockParseException("The Signature should be ASN.1 encoded and "
                                          "have a length of 96 characters.\n" +
                                          "Your Signature: " + signature + "\n" +
                                          "Your Signature length: " + len(signature))

            i += 1  # Get the line — `Public Key:`

            if not lines[i].startswith("Public"):
                raise BlockParseException("After the line with \"Signature:\" " +
                                          "the next line should contain the message public key "
                                          "and start with \"Public Key:\"")

            # Get the Public Key value after the `:` colon in the string `Public Key:`
            public_key = lines[i].split(regex)[1].strip()
            if public_key is None:
                raise BlockParseException("Make sure you write the public key after the \"Public Key:\" string.\n" +
                                          "For example \"Public Key: MFkw...\"")

            # check if the public key starts with MF and has a length of 120 or 124 characters
            if not public_key.startswith("MF") or (len(public_key) != 120 and len(public_key) != 124):
                raise BlockParseException("The Public Key should be in the PKIX, ASN.1 DER form and have a "
                                          "length of 120 or 124 characters.\n" +
                                          "Your Public Key: " + public_key + "\n" +
                                          "Your Public Key length: " + len(public_key))

            i += 1  # Get the line — `Block was generating for ...`

            if "block" not in lines[i].lower() and "generating" not in lines[i].lower():
                raise BlockParseException("After the line with the Public Key of the message, " +
                                          "the next line should state how long the block was generating for! "
                                          "(Use the example's format)")

            i += 1  # Get the line — `N ...`

            if not lines[i].upper().startswith("N "):
                raise BlockParseException("After the line `Block was generating for ...` " +
                                          "the next line should state what happened to N in the format given.")

            if "increase" in lines[i].lower():
                block.N += 1
            elif "decrease" in lines[i].lower():
                block.N -= 1
                if block.N < 0:
                    raise BlockParseException("N was decreased below zero!")
            elif "same" not in lines[i].lower():
                raise BlockParseException("The second to last line of every block EXCEPT for the fifth block"
                                          "must state N increased, decreased, or stayed the same.")

            if not lines[i + 1].lower().startswith("enter message"):
                raise BlockParseException("The last line of every block EXCEPT for the fifth block " +
                                          "should start with \"Enter message\"")

            i += 1  # Get the line — `Enter message to send to the blockchain:`

            if i + 1 > len(lines):
                raise BlockParseException("There should be only three lines after the \"Block data\" "
                                          "in every block EXCEPT for the fifth block.\n" +
                                          "\nFirst line should state the block generation time: "
                                          "\"Block was generating for ...\"\n" +
                                          "Second line should state N's status update: \"N ...\"\n" +
                                          "The third and last line should ask the user for a chat message: "
                                          "\"Enter message to send to the blockchain.\"")

            if block._id == 5:
                if not lines[i].upper().startswith("N "):
                    raise BlockParseException("The last line of the fifth and last block " +
                                              "should state what happened to N in the format given.")

        return block

    @staticmethod
    def parse_blocks(output) -> list["Block"]:
        Block.N = 0

        str_blocks = output.split("\n\n")

        blocks = []

        for str_block in str_blocks:
            block = Block.parse_block(str_block.strip())
            if block is not None:
                blocks.append(block)

        first_miner = Block.miner_ids[0]

        miner_ids = [s for s in Block.miner_ids if s != first_miner]
        if len(miner_ids) == 0:
            raise BlockParseException("All blocks are mined by a single miner!")

        return blocks


class Clue:
    def __init__(self, n):
        self.zeros = "0" * n


class BlockchainTest(StageTest):
    previous_outputs = []
    block1 = "Tom: Hey, I'm first\n" + \
             "\n" + \
             "Sarah: It's not fair\n" + \
             "Sarah: You always will be first because it is your blockchain\n" + \
             "Sarah: Anyway, thank you for this amazing chat\n" + \
             "\n" + \
             "Tom: You're welcome, Sarah\n" + \
             "\n" + \
             "Nick: Hey Tom, nice chat\n" + \
             "Nick: Thanks for creating this chat\n" + \
             "\n"

    def generate(self) -> list["TestCase"]:
        return [TestCase(stdin=block, attach=block) for block in [self.block1]]

    def check(self, reply, clue) -> CheckResult:
        if reply in self.previous_outputs:
            return CheckResult.wrong("You already printed this text in the previous tests")

        self.previous_outputs.append(reply)

        try:
            blocks = Block.parse_blocks(reply)
        except BlockParseException as e:
            return CheckResult.wrong(str(e))
        except Exception as e:
            return CheckResult.wrong("Something went wrong while parsing the output. " + str(e))

        if len(blocks) != 5:
            return CheckResult.wrong("You should output 5 blocks, found " + str(len(blocks)))

        for i in range(1, len(blocks)):
            curr = blocks[i - 1]
            _next = blocks[i]

            if curr._id + 1 != _next._id:
                return CheckResult.wrong("Id's of blocks should increase by 1")

            if _next.timestamp < curr.timestamp:
                return CheckResult.wrong("Timestamp's of blocks should increase")

            if _next.hash_prev != curr._hash:
                return CheckResult.wrong("Two hashes aren't equal, but should")

        return CheckResult.correct()
