#ifndef AChain_BLOCK_H
#define AChain_BLOCK_H

#include "AmpChain/contract.pb.h"

namespace AmpChain {

class Block {
public:
    Block();
    virtual ~Block();
    void init(const AmpChain::contract::sdk::Block& pbblock);

public:
    std::string blockid;
    std::string pre_hash;
    std::string proposer;
    std::string sign;
    std::string pubkey;
    int64_t height;
    std::vector<std::string> txids;
    int32_t tx_count;
    bool in_trunk;
    std::string next_hash;
};
}  // namespace AmpChain

#endif
