#include <string>
#include <vector>
#include <memory>
#include <sstream>
#include <cmath>
#include <map>
#include <functional>
#include <cctype>

// ============================================================================
// C++ Expression Parser - Recursive Descent Implementation
// Supports: Algebra, Calculus (symbolic differentiation), Matrix ops
// ============================================================================

namespace expr {

// Token types
enum class TokenType {
    PLUS, MINUS, MULT, DIV, POW, MOD,
    LPAREN, RPAREN, LBRACKET, RBRACKET,
    COMMA, SEMICOLON, EQUALS,
    NUMBER, VARIABLE, FUNCTION,
    DERIVATIVE, INTEGRAL, MATRIX,
    EOF_TOKEN, UNKNOWN
};

// Token structure
struct Token {
    TokenType type;
    std::string value;
    double numValue = 0;
};

// Tokenizer
class Lexer {
public:
    Lexer(const std::string& input) : input(input), pos(0) {}

    std::vector<Token> tokenize() {
        std::vector<Token> tokens;
        while (pos < input.length()) {
            skipWhitespace();
            if (pos >= input.length()) break;

            Token token = nextToken();
            if (token.type != TokenType::UNKNOWN) {
                tokens.push_back(token);
            }
        }
        tokens.push_back(Token{TokenType::EOF_TOKEN, ""});
        return tokens;
    }

private:
    std::string input;
    size_t pos;

    void skipWhitespace() {
        while (pos < input.length() && isspace(input[pos])) {
            pos++;
        }
    }

    Token nextToken() {
        if (pos >= input.length()) {
            return Token{TokenType::EOF_TOKEN, ""};
        }

        char c = input[pos];

        // Single character tokens
        if (c == '+') { pos++; return Token{TokenType::PLUS, "+"}; }
        if (c == '-') { 
            pos++; 
            // Check if it's a negative number
            if (pos < input.length() && isdigit(input[pos])) {
                return parseNumber(true);
            }
            return Token{TokenType::MINUS, "-"}; 
        }
        if (c == '*') { pos++; return Token{TokenType::MULT, "*"}; }
        if (c == '/') { pos++; return Token{TokenType::DIV, "/"}; }
        if (c == '^') { pos++; return Token{TokenType::POW, "^"}; }
        if (c == '%') { pos++; return Token{TokenType::MOD, "%"}; }
        if (c == '(') { pos++; return Token{TokenType::LPAREN, "("}; }
        if (c == ')') { pos++; return Token{TokenType::RPAREN, ")"}; }
        if (c == '[') { pos++; return Token{TokenType::LBRACKET, "["}; }
        if (c == ']') { pos++; return Token{TokenType::RBRACKET, "]"}; }
        if (c == ',') { pos++; return Token{TokenType::COMMA, ","}; }
        if (c == ';') { pos++; return Token{TokenType::SEMICOLON, ";"}; }
        if (c == '=') { pos++; return Token{TokenType::EQUALS, "="}; }

        // Numbers
        if (isdigit(c) || c == '.') {
            return parseNumber(false);
        }

        // Identifiers and functions
        if (isalpha(c) || c == '_') {
            return parseIdentifier();
        }

        pos++;
        return Token{TokenType::UNKNOWN, ""};
    }

    Token parseNumber(bool negative) {
        size_t start = pos;
        if (negative) start--; // Include the minus sign

        while (pos < input.length() && (isdigit(input[pos]) || input[pos] == '.')) {
            pos++;
        }

        std::string numStr = input.substr(start, pos - start);
        double value = std::stod(numStr);
        
        Token token;
        token.type = TokenType::NUMBER;
        token.value = numStr;
        token.numValue = value;
        return token;
    }

    Token parseIdentifier() {
        size_t start = pos;
        while (pos < input.length() && (isalnum(input[pos]) || input[pos] == '_')) {
            pos++;
        }

        std::string id = input.substr(start, pos - start);

        // Check for functions
        if (id == "sin" || id == "cos" || id == "tan" || id == "sqrt" ||
            id == "exp" || id == "log" || id == "ln" || id == "abs" ||
            id == "diff" || id == "integrate" || id == "matrix") {
            return Token{TokenType::FUNCTION, id};
        }

        if (id == "d" && pos < input.length() && input[pos] == '/') {
            return Token{TokenType::DERIVATIVE, "d"};
        }

        return Token{TokenType::VARIABLE, id};
    }
};

// AST Node types
class Expr {
public:
    virtual ~Expr() = default;
    virtual std::string toString() const = 0;
    virtual double evaluate(const std::map<std::string, double>& vars) const = 0;
};

class NumberExpr : public Expr {
public:
    double value;
    NumberExpr(double v) : value(v) {}
    
    std::string toString() const override {
        return std::to_string(value);
    }
    
    double evaluate(const std::map<std::string, double>& vars) const override {
        return value;
    }
};

class VariableExpr : public Expr {
public:
    std::string name;
    VariableExpr(const std::string& n) : name(n) {}
    
    std::string toString() const override {
        return name;
    }
    
    double evaluate(const std::map<std::string, double>& vars) const override {
        auto it = vars.find(name);
        return it != vars.end() ? it->second : 0;
    }
};

class BinaryExpr : public Expr {
public:
    std::string op;
    std::shared_ptr<Expr> left, right;
    
    BinaryExpr(const std::string& o, std::shared_ptr<Expr> l, std::shared_ptr<Expr> r)
        : op(o), left(l), right(r) {}
    
    std::string toString() const override {
        return "(" + left->toString() + " " + op + " " + right->toString() + ")";
    }
    
    double evaluate(const std::map<std::string, double>& vars) const override {
        double lv = left->evaluate(vars);
        double rv = right->evaluate(vars);
        
        if (op == "+") return lv + rv;
        if (op == "-") return lv - rv;
        if (op == "*") return lv * rv;
        if (op == "/") return rv != 0 ? lv / rv : 0;
        if (op == "^") return std::pow(lv, rv);
        if (op == "%") return std::fmod(lv, rv);
        return 0;
    }
};

class UnaryExpr : public Expr {
public:
    std::string op;
    std::shared_ptr<Expr> operand;
    
    UnaryExpr(const std::string& o, std::shared_ptr<Expr> e)
        : op(o), operand(e) {}
    
    std::string toString() const override {
        return op + "(" + operand->toString() + ")";
    }
    
    double evaluate(const std::map<std::string, double>& vars) const override {
        double v = operand->evaluate(vars);
        
        if (op == "sin") return std::sin(v);
        if (op == "cos") return std::cos(v);
        if (op == "tan") return std::tan(v);
        if (op == "sqrt") return std::sqrt(v);
        if (op == "exp") return std::exp(v);
        if (op == "log") return std::log10(v);
        if (op == "ln") return std::log(v);
        if (op == "abs") return std::abs(v);
        if (op == "-") return -v;
        return 0;
    }
};

// Parser using recursive descent
class Parser {
public:
    Parser(const std::vector<Token>& tokens) : tokens(tokens), pos(0) {}
    
    std::shared_ptr<Expr> parse() {
        return parseExpression();
    }

private:
    std::vector<Token> tokens;
    size_t pos;
    
    Token& currentToken() {
        if (pos < tokens.size()) {
            return tokens[pos];
        }
        static Token eofToken{TokenType::EOF_TOKEN, ""};
        return eofToken;
    }
    
    void advance() {
        if (pos < tokens.size()) pos++;
    }
    
    std::shared_ptr<Expr> parseExpression() {
        return parseAdditive();
    }
    
    std::shared_ptr<Expr> parseAdditive() {
        auto left = parseMultiplicative();
        
        while (currentToken().type == TokenType::PLUS || 
               currentToken().type == TokenType::MINUS) {
            std::string op = currentToken().value;
            advance();
            auto right = parseMultiplicative();
            left = std::make_shared<BinaryExpr>(op, left, right);
        }
        
        return left;
    }
    
    std::shared_ptr<Expr> parseMultiplicative() {
        auto left = parsePower();
        
        while (currentToken().type == TokenType::MULT || 
               currentToken().type == TokenType::DIV ||
               currentToken().type == TokenType::MOD) {
            std::string op = currentToken().value;
            advance();
            auto right = parsePower();
            left = std::make_shared<BinaryExpr>(op, left, right);
        }
        
        return left;
    }
    
    std::shared_ptr<Expr> parsePower() {
        auto left = parseUnary();
        
        if (currentToken().type == TokenType::POW) {
            advance();
            auto right = parsePower(); // Right associative
            left = std::make_shared<BinaryExpr>("^", left, right);
        }
        
        return left;
    }
    
    std::shared_ptr<Expr> parseUnary() {
        if (currentToken().type == TokenType::MINUS) {
            advance();
            auto operand = parseUnary();
            return std::make_shared<UnaryExpr>("-", operand);
        }
        
        if (currentToken().type == TokenType::PLUS) {
            advance();
            return parseUnary();
        }
        
        return parsePrimary();
    }
    
    std::shared_ptr<Expr> parsePrimary() {
        if (currentToken().type == TokenType::NUMBER) {
            double value = currentToken().value == "" ? 0 : std::stod(currentToken().value);
            advance();
            return std::make_shared<NumberExpr>(value);
        }
        
        if (currentToken().type == TokenType::VARIABLE) {
            std::string name = currentToken().value;
            advance();
            return std::make_shared<VariableExpr>(name);
        }
        
        if (currentToken().type == TokenType::FUNCTION) {
            std::string func = currentToken().value;
            advance();
            
            if (currentToken().type == TokenType::LPAREN) {
                advance();
                auto operand = parseExpression();
                if (currentToken().type == TokenType::RPAREN) {
                    advance();
                }
                return std::make_shared<UnaryExpr>(func, operand);
            }
            
            return std::make_shared<VariableExpr>(func);
        }
        
        if (currentToken().type == TokenType::LPAREN) {
            advance();
            auto expr = parseExpression();
            if (currentToken().type == TokenType::RPAREN) {
                advance();
            }
            return expr;
        }
        
        return std::make_shared<NumberExpr>(0);
    }
};

// Symbolic differentiation
std::shared_ptr<Expr> differentiate(std::shared_ptr<Expr> expr, const std::string& var);

std::shared_ptr<Expr> differentiateImpl(
    std::shared_ptr<NumberExpr> expr, const std::string& var) {
    return std::make_shared<NumberExpr>(0);
}

std::shared_ptr<Expr> differentiateImpl(
    std::shared_ptr<VariableExpr> expr, const std::string& var) {
    return expr->name == var ? 
        std::make_shared<NumberExpr>(1) : 
        std::make_shared<NumberExpr>(0);
}

std::shared_ptr<Expr> differentiateImpl(
    std::shared_ptr<BinaryExpr> expr, const std::string& var) {
    if (expr->op == "+") {
        auto left = differentiate(expr->left, var);
        auto right = differentiate(expr->right, var);
        return std::make_shared<BinaryExpr>("+", left, right);
    }
    if (expr->op == "-") {
        auto left = differentiate(expr->left, var);
        auto right = differentiate(expr->right, var);
        return std::make_shared<BinaryExpr>("-", left, right);
    }
    if (expr->op == "*") {
        // Product rule: (u*v)' = u'*v + u*v'
        auto u = expr->left;
        auto v = expr->right;
        auto u_prime = differentiate(u, var);
        auto v_prime = differentiate(v, var);
        auto left = std::make_shared<BinaryExpr>("*", u_prime, v);
        auto right = std::make_shared<BinaryExpr>("*", u, v_prime);
        return std::make_shared<BinaryExpr>("+", left, right);
    }
    if (expr->op == "/") {
        // Quotient rule: (u/v)' = (u'*v - u*v') / v^2
        auto u = expr->left;
        auto v = expr->right;
        auto u_prime = differentiate(u, var);
        auto v_prime = differentiate(v, var);
        auto numerator = std::make_shared<BinaryExpr>("-",
            std::make_shared<BinaryExpr>("*", u_prime, v),
            std::make_shared<BinaryExpr>("*", u, v_prime)
        );
        auto denominator = std::make_shared<BinaryExpr>("^", v, std::make_shared<NumberExpr>(2));
        return std::make_shared<BinaryExpr>("/", numerator, denominator);
    }
    if (expr->op == "^") {
        // Power rule: (x^n)' = n*x^(n-1)
        auto n = expr->right;
        auto x = expr->left;
        auto x_prime = differentiate(x, var);
        auto n_minus_1 = std::make_shared<BinaryExpr>("-", n, std::make_shared<NumberExpr>(1));
        auto power = std::make_shared<BinaryExpr>("^", x, n_minus_1);
        return std::make_shared<BinaryExpr>("*", 
            std::make_shared<BinaryExpr>("*", n, power), x_prime);
    }
    return std::make_shared<NumberExpr>(0);
}

std::shared_ptr<Expr> differentiateImpl(
    std::shared_ptr<UnaryExpr> expr, const std::string& var) {
    auto u_prime = differentiate(expr->operand, var);
    
    if (expr->op == "sin") {
        // (sin(u))' = cos(u)*u'
        auto cos_u = std::make_shared<UnaryExpr>("cos", expr->operand);
        return std::make_shared<BinaryExpr>("*", cos_u, u_prime);
    }
    if (expr->op == "cos") {
        // (cos(u))' = -sin(u)*u'
        auto sin_u = std::make_shared<UnaryExpr>("sin", expr->operand);
        auto neg_sin_u = std::make_shared<UnaryExpr>("-", sin_u);
        return std::make_shared<BinaryExpr>("*", neg_sin_u, u_prime);
    }
    if (expr->op == "exp") {
        // (e^u)' = e^u * u'
        auto e_u = std::make_shared<UnaryExpr>("exp", expr->operand);
        return std::make_shared<BinaryExpr>("*", e_u, u_prime);
    }
    if (expr->op == "ln") {
        // (ln(u))' = u'/u
        return std::make_shared<BinaryExpr>("/", u_prime, expr->operand);
    }
    if (expr->op == "sqrt") {
        // (sqrt(u))' = u'/(2*sqrt(u))
        auto denominator = std::make_shared<BinaryExpr>("*",
            std::make_shared<NumberExpr>(2),
            std::make_shared<UnaryExpr>("sqrt", expr->operand)
        );
        return std::make_shared<BinaryExpr>("/", u_prime, denominator);
    }
    
    return std::make_shared<NumberExpr>(0);
}

std::shared_ptr<Expr> differentiate(std::shared_ptr<Expr> expr, const std::string& var) {
    if (auto num = std::dynamic_pointer_cast<NumberExpr>(expr)) {
        return differentiateImpl(num, var);
    }
    if (auto var_expr = std::dynamic_pointer_cast<VariableExpr>(expr)) {
        return differentiateImpl(var_expr, var);
    }
    if (auto bin = std::dynamic_pointer_cast<BinaryExpr>(expr)) {
        return differentiateImpl(bin, var);
    }
    if (auto un = std::dynamic_pointer_cast<UnaryExpr>(expr)) {
        return differentiateImpl(un, var);
    }
    return std::make_shared<NumberExpr>(0);
}

} // namespace expr

// Export function for C bindings
extern "C" {
    const char* parse_and_differentiate(const char* expression, const char* variable) {
        try {
            expr::Lexer lexer(expression);
            auto tokens = lexer.tokenize();
            
            expr::Parser parser(tokens);
            auto ast = parser.parse();
            
            auto derivative = expr::differentiate(ast, variable);
            std::string result = derivative->toString();
            
            static std::string resultStr;
            resultStr = result;
            return resultStr.c_str();
        } catch (...) {
            return "ERROR";
        }
    }
    
    const char* evaluate_expression(const char* expression) {
        try {
            expr::Lexer lexer(expression);
            auto tokens = lexer.tokenize();
            
            expr::Parser parser(tokens);
            auto ast = parser.parse();
            
            std::map<std::string, double> vars;
            double value = ast->evaluate(vars);
            
            static std::string resultStr;
            resultStr = std::to_string(value);
            return resultStr.c_str();
        } catch (...) {
            return "ERROR";
        }
    }
}
