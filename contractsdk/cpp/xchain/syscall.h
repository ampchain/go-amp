#ifndef AChain_SYSCALL_H
#define AChain_SYSCALL_H

#include <stdint.h>
#include <string>
#include <google/protobuf/message_lite.h>

namespace AmpChain {

bool syscall(const std::string& method,
             const ::google::protobuf::MessageLite& request,
             ::google::protobuf::MessageLite* response);
}

#endif
