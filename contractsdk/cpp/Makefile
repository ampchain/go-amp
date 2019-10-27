BUILD_DIR ?= build

AChain_SRCS := $(shell ls ./AmpChain/*.cc)
AChain_OBJS := $(AChain_SRCS:%=$(BUILD_DIR)/%.o)
AChain_DEPS := $(AChain_OBJS:.o=.d)

CONTRACT_SRCS := $(shell ls ./example/*.cc)
CONTRACT_BIN := $(CONTRACT_SRCS:./example/%.cc=$(BUILD_DIR)/%.wasm)

INC_DIRS := . /usr/local/include
INC_FLAGS := $(addprefix -I,$(INC_DIRS))

CPPFLAGS ?= $(INC_FLAGS) -MMD -MP -Oz -std=c++11

.PHONY: clean all

all: $(BUILD_DIR)/libAmpChain.a $(CONTRACT_BIN)

clean:
	$(RM) -r $(BUILD_DIR)

$(BUILD_DIR)/libAmpChain.a: $(AChain_OBJS)
	$(AR) rc $@ $^
	$(RANLIB) $@

# wasm target
$(BUILD_DIR)/%.wasm: $(BUILD_DIR)/example/%.cc.o $(BUILD_DIR)/libAmpChain.a
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -o $@ $^ -Oz -s ERROR_ON_UNDEFINED_SYMBOLS=0 -s DETERMINISTIC=1 -L/usr/local/lib -lprotobuf-lite -lpthread

# c++ source
$(BUILD_DIR)/%.cc.o: %.cc
	$(MKDIR_P) $(dir $@)
	$(CXX) $(CPPFLAGS) $(CXXFLAGS) -c $< -o $@

-include $(AChain_DEPS)

MKDIR_P ?= mkdir -p
