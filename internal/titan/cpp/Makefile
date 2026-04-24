CXX = g++
CXXFLAGS = -std=c++17 -O3 -march=native -fPIC -Wall -I. -I./neural
AR = ar
ARFLAGS = rcs

TARGET = libtitan.a
SRC = titan_engine.cpp
OBJ = titan_engine.o

# Neural headers (header-only implementation for now)
NEURAL_HEADERS = neural/sovereign_neural_config.hpp \
                 neural/sovereign_neural_mla.hpp \
                 neural/sovereign_neural_moe.hpp \
                 neural/sovereign_neural_transformer.hpp \
                 neural/sovereign_neural_tokenizer.hpp

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJ)
	$(AR) $(ARFLAGS) $@ $^

$(OBJ): $(SRC) titan_engine.h $(NEURAL_HEADERS)
	$(CXX) $(CXXFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJ) $(TARGET)
