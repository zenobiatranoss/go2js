// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <cctype>
#include <cstring>
#include <unordered_set>

namespace jm {

namespace {

const std::unordered_set<std::string> kw = {
    "break",    "case",   "catch",  "class",      "const",  "continue", "debugger", "default", "delete",
    "do",       "else",   "enum",   "export",     "extends", "false",   "finally",  "for",     "function",
    "if",       "import", "in",     "instanceof", "new",    "null",     "return",   "super",   "switch",
    "this",     "throw",  "true",   "try",        "typeof", "var",      "void",     "while",   "with",
    "yield",    "await",
};

const char* const pt[] = {
    ">>>=", "...", "===", "!==", "**=", "<<=", ">>=", ">>>", "&&=", "||=", "?\?=", "=>", "==", "!=", "<=", ">=",
    "&&",   "||",  "?\?",  "?.",  "++",  "--",  "+=",  "-=",  "*=",  "/=",  "%=",  "&=", "|=", "^=", "<<", ">>", "**",
};

bool is(char c) {
    unsigned char u = c;
    return std::isalpha(u) || c == '_' || c == '$' || c == '\\' || u >= 0x80;
}

bool ip(char c) {
    return is(c) || std::isdigit(static_cast<unsigned char>(c));
}

struct Sc {
    const std::string& sr;
    size_t ps = 0;
    Ts ts;
    std::vector<char> bs;
    bool nl = false;
    std::string er;
    size_t cp = 0;
    size_t cn = 1;

    explicit Sc(const std::string& s) : sr(s) {}

    char at(size_t of = 0) const {
        return ps + of < sr.size() ? sr[ps + of] : '\0';
    }

    unsigned char ub(size_t of) const {
        return static_cast<unsigned char>(at(of));
    }

    void bd(const char* ms) {
        if (er.empty()) er = ms;
        ps = sr.size();
    }

    void ad(Ty ty, size_t fr) {
        for (; cp < fr; ++cp) {
            if (sr[cp] == '\n') ++cn;
        }
        Tk tk{ty, sr.substr(fr, ps - fr), nl};
        tk.ln = cn;
        tk.of = fr;
        ts.push_back(std::move(tk));
        nl = false;
    }

    void ws() {
        while (ps < sr.size()) {
            unsigned char c = ub(0);
            if (c == '\n' || c == '\r') {
                nl = true;
                ++ps;
            } else if (c == ' ' || c == '\t' || c == '\v' || c == '\f') {
                ++ps;
            } else if (c == 0xC2 && ub(1) == 0xA0) {
                ps += 2;
            } else if (c == 0xEF && ub(1) == 0xBB && ub(2) == 0xBF) {
                ps += 3;
            } else if (c == 0xE2 && ub(1) == 0x80 && (ub(2) == 0xA8 || ub(2) == 0xA9)) {
                nl = true;
                ps += 3;
            } else if (c == '/' && at(1) == '/') {
                while (ps < sr.size() && sr[ps] != '\n' && sr[ps] != '\r') ++ps;
            } else if (c == '/' && at(1) == '*') {
                size_t en = sr.find("*/", ps + 2);
                if (en == std::string::npos) return bd("unterminated comment");
                for (size_t ix = ps; ix < en; ++ix) {
                    if (sr[ix] == '\n' || sr[ix] == '\r') nl = true;
                }
                ps = en + 2;
            } else {
                break;
            }
        }
    }

    void id() {
        while (ps < sr.size()) {
            char c = sr[ps];
            if (ip(c) && c != '\\') {
                ++ps;
            } else if (c == '\\' && at(1) == 'u') {
                ps += 2;
                if (at() == '{') {
                    while (ps < sr.size() && sr[ps] != '}') ++ps;
                    ++ps;
                } else {
                    ps += 4;
                }
            } else {
                break;
            }
        }
    }

    void dg() {
        while (std::isdigit(static_cast<unsigned char>(at())) || at() == '_') ++ps;
    }

    void nu() {
        size_t fr = ps;
        if (at() == '0' && at(1) && std::strchr("xXoObB", at(1))) {
            ps += 2;
            while (std::isxdigit(static_cast<unsigned char>(at())) || at() == '_') ++ps;
        } else {
            dg();
            if (at() == '.') {
                ++ps;
                dg();
            }
            if ((at() == 'e' || at() == 'E') &&
                (std::isdigit(static_cast<unsigned char>(at(1))) ||
                 ((at(1) == '+' || at(1) == '-') && std::isdigit(static_cast<unsigned char>(at(2)))))) {
                ps += 2;
                dg();
            }
        }
        if (at() == 'n') ++ps;
        ad(Nu, fr);
    }

    void st(char qu) {
        size_t fr = ps++;
        while (ps < sr.size() && sr[ps] != qu) {
            if (sr[ps] == '\\') ++ps;
            ++ps;
        }
        if (ps >= sr.size()) return bd("unterminated string");
        ++ps;
        ad(St, fr);
    }

    void tp(bool fs, size_t fr) {
        while (ps < sr.size()) {
            char c = sr[ps];
            if (c == '\\') {
                ps += 2;
            } else if (c == '`') {
                ++ps;
                ad(fs ? Tf : Tt, fr);
                return;
            } else if (c == '$' && at(1) == '{') {
                ps += 2;
                bs.push_back('t');
                ad(fs ? Th : Tm, fr);
                return;
            } else {
                ++ps;
            }
        }
        bd("unterminated template");
    }

    void rg() {
        size_t fr = ps++;
        bool cl = false;
        while (ps < sr.size()) {
            char c = sr[ps];
            if (c == '\\') {
                ps += 2;
            } else if (c == '[') {
                cl = true;
                ++ps;
            } else if (c == ']') {
                cl = false;
                ++ps;
            } else if (c == '/' && !cl) {
                break;
            } else if (c == '\n' || c == '\r') {
                return bd("unterminated regular expression");
            } else {
                ++ps;
            }
        }
        if (ps >= sr.size()) return bd("unterminated regular expression");
        ++ps;
        while (ip(at())) ++ps;
        ad(Rx, fr);
    }

    void pu() {
        size_t fr = ps;
        for (const char* pp : pt) {
            size_t ln = std::strlen(pp);
            if (sr.compare(ps, ln, pp) != 0) continue;
            if (ln == 2 && pp[0] == '?' && pp[1] == '.' && std::isdigit(static_cast<unsigned char>(at(2)))) continue;
            ps += ln;
            ad(Pu, fr);
            return;
        }
        char c = sr[ps++];
        if (c == '{') bs.push_back('b');
        if (c == '}' && !bs.empty()) bs.pop_back();
        ad(Pu, fr);
    }

    bool ra() const {
        if (ts.empty()) return true;
        const Tk& pv = ts.back();
        switch (pv.ty) {
            case Id:
            case Nu:
            case St:
            case Rx:
            case Tt:
            case Tf:
                return false;
            case Kw:
                return !(pv.tx == "this" || pv.tx == "super" || pv.tx == "null" || pv.tx == "true" || pv.tx == "false");
            case Pu:
                return !(pv.tx == ")" || pv.tx == "]" || pv.tx == "}" || pv.tx == "++" || pv.tx == "--");
            default:
                return true;
        }
    }

    bool dt() const {
        return !ts.empty() && ts.back().ty == Pu && (ts.back().tx == "." || ts.back().tx == "?.");
    }

    Ts go() {
        if (sr.compare(0, 2, "#!") == 0) {
            while (ps < sr.size() && sr[ps] != '\n' && sr[ps] != '\r') ++ps;
            ad(Hb, 0);
        }
        for (;;) {
            ws();
            if (ps >= sr.size()) break;
            char c = sr[ps];
            if (c == '`') {
                size_t fr = ps++;
                tp(true, fr);
            } else if (c == '}' && !bs.empty() && bs.back() == 't') {
                size_t fr = ps++;
                bs.pop_back();
                tp(false, fr);
            } else if (c == '"' || c == '\'') {
                st(c);
            } else if (std::isdigit(static_cast<unsigned char>(c)) || (c == '.' && std::isdigit(static_cast<unsigned char>(at(1))))) {
                nu();
            } else if (is(c)) {
                size_t fr = ps;
                id();
                bool kd = kw.count(sr.substr(fr, ps - fr)) && !dt();
                ad(kd ? Kw : Id, fr);
            } else if (c == '#' && is(at(1))) {
                size_t fr = ps++;
                id();
                ad(Id, fr);
            } else if (c == '/' && ra()) {
                rg();
            } else {
                pu();
            }
        }
        return ts;
    }
};

}

bool rw(const std::string& wd) {
    return kw.count(wd) > 0;
}

Ts lx(const std::string& sr, std::string* er) {
    Sc sc(sr);
    Ts ts = sc.go();
    if (!sc.er.empty()) {
        if (er) *er = sc.er;
        return Ts();
    }
    return ts;
}

}
