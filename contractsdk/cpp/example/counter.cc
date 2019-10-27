#include "AmpChain/AmpChain.h"

struct Counter : public AmpChain::Contract {};

DEFINE_METHOD(Counter, initialize) {
    AmpChain::Context* ctx = self.context();
    const std::string& creator = ctx->arg("creator");
    if (creator.empty()) {
        ctx->error("missing creator");
        return;
    }
    ctx->put_object("creator", creator);
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Counter, increase) {
    AmpChain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    ctx->get_object(key, &value);
    int cnt = 0;
    cnt = atoi(value.c_str());
    char buf[32];
    snprintf(buf, 32, "%d", cnt + 1);
    ctx->put_object(key, buf);
    ctx->ok(buf);
}

DEFINE_METHOD(Counter, get) {
    AmpChain::Context* ctx = self.context();
    const std::string& key = ctx->arg("key");
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}
