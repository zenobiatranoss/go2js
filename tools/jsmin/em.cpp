// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <unordered_set>

namespace jm {

namespace {

const std::unordered_set<std::string> rs = {"return", "break", "continue", "throw", "yield"};
const std::unordered_set<std::string> vl = {"this", "super", "null", "true", "false"};
const std::unordered_set<std::string> cn = {"in", "instanceof", "else", "catch", "finally"};

bool wd(const Tk& tk) {
    return tk.ty == Id || tk.ty == Kw || tk.ty == Nu;
}

bool ed(const Tk& tk) {
    switch (tk.ty) {
        case Id:
        case Nu:
        case St:
        case Rx:
        case Tt:
        case Tf:
            return true;
        case Kw:
            return vl.count(tk.tx) > 0;
        case Pu:
            return tk.tx == ")" || tk.tx == "]" || tk.tx == "}" || tk.tx == "++" || tk.tx == "--";
        default:
            return false;
    }
}

bool nq(const Tk& pv, const Tk& cu) {
    if (pv.ty == Kw && rs.count(pv.tx)) {
        return !(cu.ty == Pu && (cu.tx == ";" || cu.tx == "}"));
    }
    if (!ed(pv)) return false;
    switch (cu.ty) {
        case Id:
        case Nu:
        case St:
            return true;
        case Kw:
            return cn.count(cu.tx) == 0;
        case Pu:
            return cu.tx == "{" || cu.tx == "!" || cu.tx == "~" || cu.tx == "++" || cu.tx == "--" || cu.tx == "=>";
        default:
            return false;
    }
}

bool sp(const Tk& pv, const Tk& cu) {
    if (wd(pv) && wd(cu)) return true;
    if (pv.ty == Rx && wd(cu)) return true;
    char xa = pv.tx.back();
    char ya = cu.tx.front();
    if (pv.ty == Pu && cu.ty == Pu && (xa == '+' || xa == '-') && xa == ya) return true;
    if (pv.ty == Pu && xa == '/' && (ya == '/' || ya == '*')) return true;
    if (pv.ty == Pu && xa == '<' && ya == '!') return true;
    if (pv.ty == Pu && pv.tx == "?" && cu.ty == Nu && ya == '.') return true;
    if (pv.ty == Nu && cu.ty == Pu && cu.tx == "." && pv.tx.find_first_of(".eExXnbBoO") == std::string::npos) return true;
    return false;
}

}

std::string em(const Ts& ts) {
    std::string ot;
    for (size_t ix = 0; ix < ts.size(); ++ix) {
        const Tk& cu = ts[ix];
        if (ix) {
            const Tk& pv = ts[ix - 1];
            if (pv.ty == Hb || (cu.nl && nq(pv, cu))) {
                ot += '\n';
            } else if (sp(pv, cu)) {
                ot += ' ';
            }
        }
        ot += cu.tx;
    }
    return ot;
}

bool ok(const Ts& a, const Ts& b) {
    if (a.size() != b.size()) return false;
    for (size_t ix = 0; ix < a.size(); ++ix) {
        if (a[ix].ty != b[ix].ty || a[ix].tx != b[ix].tx) return false;
    }
    return true;
}

}
