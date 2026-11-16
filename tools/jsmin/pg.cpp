// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <algorithm>
#include <map>
#include <set>
#include <unordered_map>
#include <unordered_set>

namespace jm {

namespace {

bool pu(const Tk& tk, const char* sr) {
    return tk.ty == Pu && tk.tx == sr;
}

bool wo(const Tk& tk, const char* sr) {
    return tk.ty == Kw && tk.tx == sr;
}

std::map<std::string, char> dm(const Ts& ts, size_t bi, size_t ei) {
    std::map<std::string, char> dc;
    size_t dp = 0;
    for (size_t ix = bi; ix < ei; ++ix) {
        const Tk& tk = ts[ix];
        if (tk.ty == Pu && (tk.tx == "(" || tk.tx == "[" || tk.tx == "{")) ++dp;
        else if (tk.ty == Th) ++dp;
        else if (tk.ty == Tt && dp) --dp;
        else if (tk.ty == Pu && (tk.tx == ")" || tk.tx == "]" || tk.tx == "}") && dp) --dp;
        if (dp != 0 || ix + 1 >= ei) continue;
        bool pv = ix == bi || (ts[ix - 1].ty == Pu && (ts[ix - 1].tx == ";" || ts[ix - 1].tx == "}" || ts[ix - 1].tx == "{")) || ts[ix - 1].ty == Hb;
        if (tk.ty == Kw && tk.tx == "function") {
            size_t jx = ix + 1;
            if (pu(ts[jx], "*") && jx + 1 < ei) ++jx;
            if (ts[jx].ty == Id) dc.emplace(ts[jx].tx, 'f');
        } else if (tk.ty == Kw && tk.tx == "class" && ts[ix + 1].ty == Id) {
            dc.emplace(ts[ix + 1].tx, 'c');
        } else if (pv && ((tk.ty == Kw && (tk.tx == "const" || tk.tx == "var")) || (tk.ty == Id && tk.tx == "let" && ts[ix + 1].ty != Pu))) {
            char kd = tk.tx == "const" ? 'k' : 'v';
            size_t jx = ix + 1;
            size_t cd = 0;
            bool nm = true;
            for (; jx < ei; ++jx) {
                const Tk& cu = ts[jx];
                bool op = cu.ty == Pu && (cu.tx == "(" || cu.tx == "[" || cu.tx == "{");
                bool cl = cu.ty == Pu && (cu.tx == ")" || cu.tx == "]" || cu.tx == "}");
                if (cd == 0 && pu(cu, ";")) break;
                if (cd == 0 && jx > ix + 1 && cu.nl && !nm && nq(ts[jx - 1], cu)) break;
                if (cu.ty == Th) ++cd;
                if (cu.ty == Tt && cd) --cd;
                if (op) ++cd;
                if (cl && cd) --cd;
                if (cd == 0 && pu(cu, ",")) {
                    nm = true;
                    continue;
                }
                if (nm && cu.ty == Id) {
                    dc.emplace(cu.tx, kd);
                    nm = false;
                } else if (nm && (pu(cu, "[") || pu(cu, "{"))) {
                    size_t dd = 1;
                    for (size_t kx = jx + 1; kx < ei && dd; ++kx) {
                        const Tk& pt = ts[kx];
                        if (pt.ty == Pu && (pt.tx == "[" || pt.tx == "{")) ++dd;
                        if (pt.ty == Pu && (pt.tx == "]" || pt.tx == "}")) --dd;
                        if (pt.ty == Id && !(kx + 1 < ei && pu(ts[kx + 1], ":")) && !(kx > 0 && pu(ts[kx - 1], "."))) dc.emplace(pt.tx, kd);
                    }
                    nm = false;
                }
            }
            ix = jx > ix ? jx - 1 : ix;
            dp = 0;
        }
    }
    return dc;
}

void yp(Ts& ts) {
    const size_t no = std::string::npos;
    std::vector<size_t> mt(ts.size(), no);
    std::vector<size_t> sk;
    for (size_t ix = 0; ix < ts.size(); ++ix) {
        if (pu(ts[ix], "(")) {
            sk.push_back(ix);
        } else if (pu(ts[ix], ")") && !sk.empty()) {
            mt[sk.back()] = ix;
            sk.pop_back();
        }
    }

    std::vector<bool> dr(ts.size(), false);
    for (size_t ix = 0; ix + 1 < ts.size(); ++ix) {
        if (!pu(ts[ix], "(") || !wo(ts[ix + 1], "yield") || mt[ix] == no || mt[ix] + 1 >= ts.size()) continue;
        if (ix == 0) continue;
        const Tk& pv = ts[ix - 1];
        const Tk& nx = ts[mt[ix] + 1];
        bool op = (pv.ty == Pu && (pv.tx == "(" || pv.tx == "," || pv.tx == "=" || pv.tx == ";" || pv.tx == "{" || pv.tx == "}" || pv.tx == "[")) ||
                  wo(pv, "return");
        bool on = nx.ty == Pu && (nx.tx == "," || nx.tx == ")" || nx.tx == ";" || nx.tx == "]" || nx.tx == "}");
        if (!op || !on) continue;
        size_t dp = 0;
        bool cm = false;
        for (size_t jx = ix + 1; jx < mt[ix]; ++jx) {
            const Tk& tk = ts[jx];
            if (tk.ty == Pu && (tk.tx == "(" || tk.tx == "[" || tk.tx == "{")) ++dp;
            if (tk.ty == Pu && (tk.tx == ")" || tk.tx == "]" || tk.tx == "}")) --dp;
            if (tk.ty == Th) ++dp;
            if (tk.ty == Tt) --dp;
            if (dp == 0 && pu(tk, ",")) cm = true;
        }
        if (cm) continue;
        dr[ix] = true;
        dr[mt[ix]] = true;
    }

    Ts ou;
    ou.reserve(ts.size());
    bool cr = false;
    for (size_t ix = 0; ix < ts.size(); ++ix) {
        if (dr[ix]) {
            cr = cr || ts[ix].nl;
            continue;
        }
        ou.push_back(ts[ix]);
        if (cr) {
            ou.back().nl = true;
            cr = false;
        }
    }
    ts.swap(ou);
}

void tn(Ts& ts) {
    std::unordered_set<std::string> us;
    for (const Tk& tk : ts) {
        if (tk.ty == Id || tk.ty == Kw) us.insert(tk.tx);
    }
    std::unordered_map<std::string, std::string> mp;
    for (Tk& tk : ts) {
        if (tk.ty != Id || tk.tx.compare(0, 7, "$go2js_") != 0) continue;
        auto it = mp.find(tk.tx);
        if (it == mp.end()) {
            std::string bs = tk.tx.substr(7);
            size_t us_ = bs.rfind('_');
            std::string nw = bs;
            if (us_ != std::string::npos && us_ + 1 < bs.size() && std::all_of(bs.begin() + static_cast<long>(us_) + 1, bs.end(), ::isdigit)) {
                nw = bs.substr(0, us_) + bs.substr(us_ + 1);
            }
            if (nw.empty() || std::isdigit(static_cast<unsigned char>(nw[0])) || rw(nw)) nw = "tmp" + nw;
            while (us.count(nw)) nw += "_";
            us.insert(nw);
            it = mp.emplace(tk.tx, nw).first;
        }
        tk.tx = it->second;
    }
}

}

void pc(Ts& ts) {
    yp(ts);
    tn(ts);
}

Vw vw(const Ts& ts) {
    Vw rs;
    size_t dp = 0;
    for (size_t ix = 0; ix < ts.size(); ++ix) {
        const Tk& tk = ts[ix];
        if (tk.ty == Pu && (tk.tx == "(" || tk.tx == "[" || tk.tx == "{")) ++dp;
        if (tk.ty == Pu && (tk.tx == ")" || tk.tx == "]" || tk.tx == "}") && dp) --dp;
        if (tk.ty == Th) ++dp;
        if (tk.ty == Tt && dp) --dp;
        if (tk.ty == Id && tk.tx == "go2jsSetPackageName" && ix + 1 < ts.size() && pu(ts[ix + 1], "(") &&
            !(ix && wo(ts[ix - 1], "function"))) {
            rs.fd = true;
            rs.bi = ix;
            rs.dp = dp;
            break;
        }
    }
    if (!rs.fd) return rs;
    rs.ei = ts.size();
    if (rs.dp > 0) {
        size_t cd = rs.dp;
        for (size_t ix = rs.bi; ix < ts.size(); ++ix) {
            const Tk& tk = ts[ix];
            if (tk.ty == Pu && (tk.tx == "(" || tk.tx == "[" || tk.tx == "{")) ++cd;
            if (tk.ty == Th) ++cd;
            if (tk.ty == Tt) --cd;
            if (tk.ty == Pu && (tk.tx == ")" || tk.tx == "]" || tk.tx == "}")) {
                --cd;
                if (cd < rs.dp) {
                    rs.ei = ix;
                    break;
                }
            }
        }
    }
    return rs;
}

Ts sc(const Ts& ts, size_t bi, size_t ei) {
    return Ts(ts.begin() + static_cast<long>(bi), ts.begin() + static_cast<long>(ei));
}

bool sx(const std::string& sr, const Ts& ts, const std::string& bs, bool es, std::string& rt, std::string& pr, std::string& er) {
    Vw vv = vw(ts);
    if (!vv.fd) {
        er = "no program found: go2jsSetPackageName is never called";
        return false;
    }
    if (vv.dp != 0) {
        er = "the program is inside a function (the iife form), which cannot be split; use -module commonjs or esm";
        return false;
    }

    std::map<std::string, char> dc = dm(ts, 0, vv.bi);
    std::map<std::string, char> pd = dm(ts, vv.bi, vv.ei);

    std::set<std::string> us, vs;
    for (size_t ix = vv.bi; ix < vv.ei; ++ix) {
        const Tk& tk = ts[ix];
        if (tk.ty != Id) continue;
        if (ix > 0 && ts[ix - 1].ty == Pu && (ts[ix - 1].tx == "." || ts[ix - 1].tx == "?.")) continue;
        if (pd.count(tk.tx) && tk.tx != "go2jsSetPackageName") continue;
        auto it = dc.find(tk.tx);
        bool rn = tk.tx.compare(0, 5, "go2js") == 0 || tk.tx.compare(0, 7, "__go2js") == 0;
        if (it == dc.end()) {
            (void)rn;
            continue;
        }
        if (it->second == 'v') vs.insert(tk.tx);
        else us.insert(tk.tx);
    }
    us.erase("go2jsSetPackageName");
    us.insert("go2jsSetPackageName");
    if (!vs.empty()) us.insert("go2jsState");

    std::string ls;
    bool fs = true;
    for (const std::string& nm : us) {
        if (!dc.count(nm) && nm != "go2jsState") continue;
        ls += std::string(fs ? "" : ",\n") + "    " + nm;
        fs = false;
    }

    std::string rf = "./" + bs + ".runtime.js";
    Ts pg = sc(ts, vv.bi, vv.ei);
    pc(pg);
    if (!vs.empty()) {
        Ts ou;
        for (size_t ix = 0; ix < pg.size(); ++ix) {
            bool dt = ix > 0 && pg[ix - 1].ty == Pu && (pg[ix - 1].tx == "." || pg[ix - 1].tx == "?.");
            if (pg[ix].ty == Id && vs.count(pg[ix].tx) && !dt) {
                Tk st = pg[ix];
                st.tx = "go2jsState";
                Tk pt = pg[ix];
                pt.ty = Pu;
                pt.tx = ".";
                pt.nl = false;
                ou.push_back(st);
                ou.push_back(pt);
                pg[ix].nl = false;
            }
            ou.push_back(pg[ix]);
        }
        pg.swap(ou);
    }
    pr = es ? "import {\n" + ls + ",\n} from \"" + rf + "\";\n\n" : "\"use strict\";\n\nconst {\n" + ls + ",\n} = require(\"" + rf + "\");\n\n";
    pr += pf(pg);

    rt = sr.substr(0, ts[vv.bi].of);
    if (!rt.empty() && rt.back() != '\n') rt += '\n';
    std::string ex;
    for (const std::string& nm : us) {
        if (dc.count(nm) || nm == "go2jsState") ex += "    " + nm + ",\n";
    }
    if (!vs.empty()) {
        rt += "\nconst go2jsState = {\n";
        for (const std::string& nm : vs) {
            rt += "    get " + nm + "() {\n        return " + nm + ";\n    },\n";
            rt += "    set " + nm + "(value) {\n        " + nm + " = value;\n    },\n";
        }
        rt += "};\n";
    }
    rt += es ? "\nexport {\n" + ex + "};\n" : "\nmodule.exports = {\n" + ex + "};\n";
    return true;
}

}
