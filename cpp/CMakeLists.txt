cmake_minimum_required(VERSION 3.21)
project(TitanEngine VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Fetch Content for Dependencies
include(FetchContent)

# Google Test
FetchContent_Declare(
  googletest
  URL https://github.com/google/googletest/archive/refs/heads/main.zip
)
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

# Google Benchmark
FetchContent_Declare(
  googlebenchmark
  URL https://github.com/google/benchmark/archive/refs/heads/main.zip
)
FetchContent_MakeAvailable(googlebenchmark)

# llama.cpp
FetchContent_Declare(
  llama
  GIT_REPOSITORY https://github.com/ggerganov/llama.cpp.git
  GIT_TAG master
)
FetchContent_MakeAvailable(llama)

# OpenMP
find_package(OpenMP REQUIRED)

# Common Compiler Flags
set(COMMON_FLAGS -Wall -Wextra -Wpedantic -Werror)

# Build Type Specific Flags
if(CMAKE_BUILD_TYPE STREQUAL "Debug")
    add_compile_options(${COMMON_FLAGS} -O0 -g -fsanitize=address,undefined)
    add_link_options(-fsanitize=address,undefined)
else()
    add_compile_options(${COMMON_FLAGS} -O3 -march=native -ffast-math)
endif()

# 1. Titan Engine Library (Shared)
add_library(titan_engine SHARED src/inference_queue.cpp src/titan.cpp)
target_include_directories(titan_engine PUBLIC include src)
target_link_libraries(titan_engine PRIVATE OpenMP::OpenMP_CXX llama)

# 2. Titan Tests
add_executable(titan_test tests/test_queue.cpp)
target_link_libraries(titan_test PRIVATE titan_engine gtest_main)
include(GoogleTest)
gtest_discover_tests(titan_test)

# 3. Titan Benchmarks
add_executable(titan_bench bench/bench_inference.cpp)
target_link_libraries(titan_bench PRIVATE titan_engine benchmark::benchmark)

# Install Rules (optional but good practice)
install(TARGETS titan_engine DESTINATION lib)
install(FILES include/titan.h DESTINATION include)
