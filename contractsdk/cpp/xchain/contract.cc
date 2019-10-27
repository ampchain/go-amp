#include "AmpChain/context_impl.h"
#include "AmpChain/contract.pb.h"
#include "AmpChain/syscall.h"
#include "AmpChain/AmpChain.h"

namespace AmpChain {

static void return_response(const Response& resp) {
    pb::SetOutputRequest req;
    pb::SetOutputResponse rep;
    pb::Response* r = req.mutable_response();
    r->set_status(resp.status);
    r->set_message(resp.message);
    r->set_body(resp.body);
    syscall("SetOutput", req, &rep);
}

Contract::Contract() {
    ContextImpl* ctximpl = new (ContextImpl);
    ctximpl->init();
    _ctx = ctximpl;
}

Contract::~Contract() {
    return_response(*_ctx->mutable_response());
    delete (_ctx);
}

}  // namespace AmpChain
