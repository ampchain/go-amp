#include "AmpChain/AmpChain.h"

struct ComplianceCheck : public AmpChain::Contract {};

DEFINE_METHOD(ComplianceCheck, initialize) {
    AmpChain::Context* ctx = self.context();
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(ComplianceCheck, call) {
    AmpChain::Context* ctx = self.context();
    ctx->ok("access permission succeed");
}

