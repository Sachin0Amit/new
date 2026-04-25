#ifndef TITAN_H
#define TITAN_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stddef.h>

/**
 * Handle representing an instance of the Titan Inference Engine.
 */
typedef struct titan_handle_t titan_handle_t;

/**
 * Error Codes
 */
#define TITAN_OK 0
#define TITAN_ERR_INIT -1
#define TITAN_ERR_INVALID_HANDLE -2
#define TITAN_ERR_QUEUE_FULL -3
#define TITAN_ERR_INFERENCE_FAILED -4

/**
 * Initialize the Titan engine with a configuration JSON string.
 * Returns a handle on success, NULL on failure.
 */
titan_handle_t* titan_init(const char* config_json);

/**
 * Run inference for a given prompt.
 * Returns a dynamically allocated string with the result, or NULL on error.
 * Caller must free the result using titan_free_result.
 */
char* titan_derive(titan_handle_t* handle, const char* prompt, int max_tokens);

/**
 * Free a string result returned by titan_derive.
 */
void titan_free_result(char* result);

/**
 * Destroy the Titan engine instance and free its resources.
 */
void titan_destroy(titan_handle_t* handle);

#ifdef __cplusplus
}
#endif

#endif // TITAN_H
