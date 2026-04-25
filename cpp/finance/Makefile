CXX = g++
CXXFLAGS = -std=c++17 -O3 -march=native -fPIC -Wall -Wextra -I.
LDFLAGS = -shared

TARGET = libfinance_engine.so
SRC = finance_engine.cpp
OBJ = finance_engine.o

.PHONY: all clean test

all: $(TARGET)

$(TARGET): $(OBJ)
	$(CXX) $(LDFLAGS) -o $@ $^

$(OBJ): $(SRC) engine.hpp types.hpp math.hpp indicators.hpp predictions.hpp risk.hpp patterns.hpp backtester.hpp sentiment.hpp fundamentals.hpp
	$(CXX) $(CXXFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJ) $(TARGET) test_engine

test: test_engine
	./test_engine

test_engine: test_engine.cpp engine.hpp
	$(CXX) -std=c++17 -O2 -I. -o $@ $<
