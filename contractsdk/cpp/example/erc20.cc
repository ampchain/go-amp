#include "AmpChain/AmpChain.h"

struct ERC20 : public AmpChain::Contract {};

const std::string BALANCEPRE = "balanceOf_";
const std::string ALLOWANCEPRE = "allowanceOf_";

DEFINE_METHOD(ERC20, initialize) {
    AmpChain::Context* ctx = self.context();
    const std::string& caller = ctx->arg("caller");
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }
    const std::string& totalSupply = ctx->arg("totalSupply");
    if (totalSupply.empty()) {
        ctx->error("missing totalSupply");
        return;
    }

    std::string key = BALANCEPRE + caller;
    ctx->put_object("totalSupply", totalSupply);
    ctx->put_object(key, totalSupply);
}

DEFINE_METHOD(ERC20, totalSupply) {
    AmpChain::Context* ctx = self.context();
    std::string value;
    if (ctx->get_object("totalSupply", &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, balance) {
    AmpChain::Context* ctx = self.context();
    const std::string& caller = ctx->arg("caller");
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }
    
    std::string key = BALANCEPRE + caller;
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, allowance) {
    AmpChain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
   
    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    std::string key = ALLOWANCEPRE + from + "_" + to;
    std::string value;
    if (ctx->get_object(key, &value)) {
        ctx->ok(value);
    } else {
        ctx->error("key not found");
    }
}

DEFINE_METHOD(ERC20, transfer) {
    AmpChain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
   
    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    const std::string& token_str = ctx->arg("token");
    if (token_str.empty()) {
        ctx->error("missing token");
        return;
    }
    int token = atoi(token_str.c_str());

    std::string from_key = BALANCEPRE + from;
    std::string value;
    int from_balance = 0;
    if (ctx->get_object(from_key, &value)) {
        from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("key not found");
        return;
    }

    std::string to_key = BALANCEPRE + to;
    int to_balance = 0;
    if (ctx->get_object(to_key, &value)) {
        to_balance = atoi(value.c_str());
    }
   
    from_balance = from_balance - token;
    to_balance = to_balance + token;
   
    char buf[32]; 
    snprintf(buf, 32, "%d", from_balance);
    ctx->put_object(from_key, buf);
    snprintf(buf, 32, "%d", to_balance);
    ctx->put_object(to_key, buf);

    ctx->ok("transfer success");
}

DEFINE_METHOD(ERC20, transferFrom) {
    AmpChain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
  
    const std::string& caller = ctx->arg("caller");
    if (caller.empty()) {
        ctx->error("missing caller");
        return;
    }

    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    const std::string& token_str = ctx->arg("token");
    if (token_str.empty()) {
        ctx->error("missing token");
        return;
    }
    int token = atoi(token_str.c_str());

    std::string allowance_key = ALLOWANCEPRE + from + "_" + caller;
    std::string value;
    int allowance_balance = 0;
    if (ctx->get_object(allowance_key, &value)) {
        allowance_balance = atoi(value.c_str()); 
        if (allowance_balance < token) {
            ctx->error("The allowance of from_to not enough");
            return;
        }  
    } else {
        ctx->error("You need to add allowance from_to");
        return;
    }

    std::string from_key = BALANCEPRE + from;
    int from_balance = 0;
    if (ctx->get_object(from_key, &value)) {
        from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("From no balance");
        return;
    }

    std::string to_key = BALANCEPRE + to;
    int to_balance = 0;
    if (ctx->get_object(to_key, &value)) {
        to_balance = atoi(value.c_str());
    }
   
    from_balance = from_balance - token;
    to_balance = to_balance + token;
    allowance_balance = allowance_balance - token;

    char buf[32]; 
    snprintf(buf, 32, "%d", from_balance);
    ctx->put_object(from_key, buf);
    snprintf(buf, 32, "%d", to_balance);
    ctx->put_object(to_key, buf);
    snprintf(buf, 32, "%d", allowance_balance);
    ctx->put_object(allowance_key, buf);

    ctx->ok("transferFrom success");
}

DEFINE_METHOD(ERC20, approve) {
    AmpChain::Context* ctx = self.context();
    const std::string& from = ctx->arg("from");
    if (from.empty()) {
        ctx->error("missing from");
        return;
    }
   
    const std::string& to = ctx->arg("to");
    if (to.empty()) {
        ctx->error("missing to");
        return;
    }

    const std::string& token_str = ctx->arg("token");
    if (token_str.empty()) {
        ctx->error("missing token");
        return;
    }
    int token = atoi(token_str.c_str());

    std::string from_key = BALANCEPRE + from;
    std::string value;
    if (ctx->get_object(from_key, &value)) {
        int from_balance = atoi(value.c_str()); 
        if (from_balance < token) {
            ctx->error("The balance of from not enough");
            return;
        }  
    } else {
        ctx->error("From no balance");
        return;
    }

    std::string allowance_key = ALLOWANCEPRE + from + "_" + to;
    int allowance_balance = 0;
    if (ctx->get_object(allowance_key, &value)) {
        allowance_balance = atoi(value.c_str()); 
    }

    allowance_balance = allowance_balance + token;
   
    char buf[32]; 
    snprintf(buf, 32, "%d", allowance_balance);
    ctx->put_object(allowance_key, buf);

    ctx->ok("approve success");
}



