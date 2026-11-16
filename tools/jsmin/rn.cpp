// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <algorithm>
#include <unordered_map>
#include <unordered_set>

namespace jm {

namespace {

const char* const hd = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$";
const char* const tl = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$0123456789";

bool ow(const std::string& nm) {
    return nm.compare(0, 5, "go2js") == 0 || nm.compare(0, 7, "__go2js") == 0;
}

bool ch(char c) {
    return std::isalnum(static_cast<unsigned char>(c)) || c == '_' || c == '$';
}

std::string nm(size_t ix) {
    std::string rs(1, hd[ix % 54]);
    ix /= 54;
    while (ix) {
        --ix;
        rs += tl[ix % 64];
        ix /= 64;
    }
    return rs;
}

bool tx(const Tk& tk) {
    return tk.ty == St || tk.ty == Th || tk.ty == Tm || tk.ty == Tt || tk.ty == Tf || tk.ty == Rx;
}

void sc(const Tk& tk, std::unordered_set<std::string>& ex, std::vector<std::string>& pf) {
    const std::string& sr = tk.tx;
    size_t ps = 0;
    while ((ps = sr.find("go2js", ps)) != std::string::npos) {
        size_t lo = ps;
        while (lo > 0 && ch(sr[lo - 1])) --lo;
        size_t hi = ps;
        while (hi < sr.size() && ch(sr[hi])) ++hi;
        std::string rq = sr.substr(lo, hi - lo);
        ex.insert(rq);
        bool en = hi >= sr.size() || sr[hi] == '`' || sr[hi] == '"' || sr[hi] == '\'' || sr[hi] == '$';
        if (en && rq != "go2js" && rq != "__go2js" && rq != "__go2js_") pf.push_back(rq);
        ps = hi;
    }
}

}

size_t rn(Ts& ts) {
    std::unordered_map<std::string, size_t> fq;
    std::unordered_set<std::string> us, ex;
    std::vector<std::string> pf;

    for (const Tk& tk : ts) {
        if (tk.ty == Id || tk.ty == Kw) us.insert(tk.tx);
        if (tk.ty == Id && ow(tk.tx)) ++fq[tk.tx];
        if (tx(tk)) sc(tk, ex, pf);
    }

    std::vector<std::pair<std::string, size_t>> cs;
    for (const auto& en : fq) {
        bool bn = ex.count(en.first) > 0;
        for (const std::string& pr : pf) {
            if (en.first.compare(0, pr.size(), pr) == 0) bn = true;
        }
        if (!bn) cs.push_back(en);
    }

    std::sort(cs.begin(), cs.end(), [](const auto& a, const auto& b) {
        size_t wa = a.second * a.first.size();
        size_t wb = b.second * b.first.size();
        return wa != wb ? wa > wb : a.first < b.first;
    });

    std::unordered_map<std::string, std::string> mp;
    size_t ia = 0, ib = 0;
    for (const auto& en : cs) {
        bool pk = en.first.compare(0, 2, "__") == 0;
        size_t& ix = pk ? ib : ia;
        std::string nw;
        do {
            nw = (pk ? "__go2js_" : "go2js") + nm(ix++);
        } while (us.count(nw));
        mp[en.first] = nw;
    }

    for (Tk& tk : ts) {
        if (tk.ty != Id) continue;
        auto it = mp.find(tk.tx);
        if (it != mp.end()) tk.tx = it->second;
    }
    return mp.size();
}

}
