// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <algorithm>
#include <cstdlib>
#include <unordered_map>
#include <unordered_set>

namespace jm {

namespace {

const size_t wd = 110;

struct Nd {
    bool gp = false;
    char op = 0;
    bool ob = false;
    Tk tk{};
    Tk cl{};
    std::vector<Nd> ch;
};

const std::unordered_set<std::string> vk = {"this", "super", "null", "true", "false"};
const std::unordered_set<std::string> sk = {"if",    "for",    "while", "switch", "catch", "with",       "return",
                                              "typeof", "void",  "delete", "throw",  "yield", "await",      "in",
                                              "instanceof", "case", "else", "do",     "new",   "of"};
const std::unordered_set<std::string> fk = {"else", "try", "finally", "do", "static", "default"};
const std::unordered_set<std::string> hk = {"if", "while", "for", "switch", "catch", "with"};

bool pn(const Tk& tk, const char* sr) {
    return tk.ty == Pu && tk.tx == sr;
}

bool kn(const Tk& tk, const char* sr) {
    return tk.ty == Kw && tk.tx == sr;
}

bool ve(const Tk& tk) {
    switch (tk.ty) {
        case Id:
        case Nu:
        case St:
        case Rx:
        case Tt:
        case Tf:
            return true;
        case Kw:
            return vk.count(tk.tx) > 0;
        case Pu:
            return tk.tx == ")" || tk.tx == "]" || tk.tx == "}";
        default:
            return false;
    }
}

const Tk& lt(const Nd& nd) {
    return nd.gp ? nd.cl : nd.tk;
}

size_t le(const Tk& tk) {
    size_t cn = tk.ln;
    for (char c : tk.tx) {
        if (c == '\n') ++cn;
    }
    return cn;
}

bool oj(const Nd* pv, const Nd& pa) {
    if (!pv) return pa.op == '(' || pa.op == '[' || pa.op == 't';
    if (pv->gp) return false;
    const Tk& tk = pv->tk;
    if (tk.ty == Pu) {
        if (tk.tx == "=>" || tk.tx == ";" || tk.tx == "{" || tk.tx == "}") return false;
        if (tk.tx == ":") return (pa.op == '{' && pa.ob) || pa.op == '(' || pa.op == '[';
        return true;
    }
    if (tk.ty == Kw) return fk.count(tk.tx) == 0;
    return false;
}

void ps(const Ts& ts, size_t& ix, Nd& pa) {
    while (ix < ts.size()) {
        const Tk& tk = ts[ix];
        bool en = (pa.op == '{' && pn(tk, "}")) || (pa.op == '(' && pn(tk, ")")) || (pa.op == '[' && pn(tk, "]")) ||
                  (pa.op == 't' && tk.ty == Tt);
        if (en) {
            pa.cl = tk;
            ++ix;
            return;
        }
        Nd nd;
        if (tk.ty == Pu && (tk.tx == "(" || tk.tx == "[" || tk.tx == "{")) {
            nd.gp = true;
            nd.op = tk.tx[0];
            nd.tk = tk;
            if (nd.op == '{') nd.ob = oj(pa.ch.empty() ? nullptr : &pa.ch.back(), pa);
            ++ix;
            ps(ts, ix, nd);
        } else if (tk.ty == Th) {
            nd.gp = true;
            nd.op = 't';
            nd.tk = tk;
            ++ix;
            ps(ts, ix, nd);
        } else {
            nd.tk = tk;
            ++ix;
        }
        pa.ch.push_back(std::move(nd));
    }
}

struct Fm {
    bool fo = false;
    bool bd = false;
    bool kc = false;
    std::vector<std::string> ls;
    std::vector<size_t> ol;
    std::string cu;
    bool bl = true;
    size_t cf = 0;
    Tk lk{};
    bool hv = false;
    bool un = false;
    int gs = 0;
    bool dl = false;

    Fm() = default;

    Fm(const Fm& o, bool) : fo(true), kc(o.kc), cu(), bl(o.bl), cf(o.cf), lk(o.lk), hv(o.hv), un(o.un), gs(o.gs) {}

    void nl(int ind) {
        if (fo) return;
        if (!bl) {
            ls.push_back(cu);
            ol.push_back(cf);
        }
        cu.assign(static_cast<size_t>(ind) * 4, ' ');
        bl = true;
        cf = 0;
    }

    void bk() {
        if (fo) return;
        if (!bl) {
            ls.push_back(cu);
            ol.push_back(cf);
        }
        cu.clear();
        bl = true;
        cf = 0;
        if (!ls.empty() && !ls.back().empty()) {
            ls.push_back("");
            ol.push_back(0);
        }
    }

    bool sg(const Tk& b) const {
        const Tk& a = lk;
        if (a.ty == Th || a.ty == Tm || b.ty == Tm || b.ty == Tt) return false;
        if (a.ty == Pu && (a.tx == "(" || a.tx == "[" || a.tx == "{" || a.tx == "." || a.tx == "?." || a.tx == "...")) return false;
        if (un && a.ty == Pu) {
            return b.ty == Pu && a.tx.back() == b.tx[0] && (b.tx[0] == '+' || b.tx[0] == '-');
        }
        bool ae = ve(a);
        if (b.ty == Pu) {
            const std::string& sr = b.tx;
            if (sr == "," || sr == ";" || sr == ")" || sr == "]" || sr == "." || sr == "?.") return false;
            if (sr == ":") return !kc;
            if (sr == "(") {
                if (a.ty == Kw && !vk.count(a.tx)) return sk.count(a.tx) > 0;
                if (a.ty == Pu && a.tx == "*" && gs) return gs == 2;
                return !ae;
            }
            if (sr == "[") return !ae || (a.ty == Id && (a.tx == "let" || a.tx == "of" || a.tx == "static" || a.tx == "async"));
            if (sr == "{") return true;
            if (sr == "*" && a.ty == Kw && (a.tx == "function" || a.tx == "yield")) return false;
            if (sr == "++" || sr == "--") return !ae;
            if (sr == "!" || sr == "~") return true;
            return true;
        }
        if ((b.ty == Tf || b.ty == Th) && ae && a.ty != Kw) return false;
        return true;
    }

    void ap(const Tk& tk) {
        if (bl) {
            cf = tk.ln;
            bl = false;
        }
        cu += tk.tx;
        lk = tk;
        hv = true;
        un = false;
        gs = 0;
    }

    void pt(const Tk& tk) {
        bool sp = !bl && hv && sg(tk);
        bool pre = false;
        int gn = 0;
        if (tk.ty == Pu) {
            bool ae = hv && ve(lk);
            if (tk.tx == "!" || tk.tx == "~") pre = true;
            else if (tk.tx == "+" || tk.tx == "-" || tk.tx == "++" || tk.tx == "--") pre = !ae;
            if (tk.tx == "*" && hv && lk.ty == Kw) gn = lk.tx == "function" ? 1 : (lk.tx == "yield" ? 2 : 0);
        }
        if (sp) cu += ' ';
        if (bl) {
            cf = tk.ln;
            bl = false;
        }
        cu += tk.tx;
        lk = tk;
        hv = true;
        un = pre;
        gs = gn;
    }

    static bool hb(const Nd& nd) {
        if (nd.gp && nd.op == '{' && !nd.ob) return true;
        if (nd.gp) {
            for (const Nd& cd : nd.ch) {
                if (hb(cd)) return true;
            }
        }
        return false;
    }

    void ib(const Nd* b, const Nd* e, int ind) {
        rr(b, e, ind, false, false);
    }

    void ex(const Nd& nd, int ind, bool nb) {
        if (fo) {
            ln(nd, ind);
            return;
        }
        Fm tf(*this, true);
        tf.ln(nd, ind);
        size_t cl = cu.size();
        if (!tf.bd && (nb || cl + tf.cu.size() + 2 <= wd)) {
            cu += tf.cu;
            if (bl && !tf.bl) cf = tf.cf;
            bl = tf.bl;
            lk = tf.lk;
            hv = tf.hv;
            un = tf.un;
            gs = tf.gs;
            return;
        }
        if (nd.op == 't' || (tf.bd && !nd.ob)) {
            ln(nd, ind);
            return;
        }
        pt(nd.tk);
        if (!nd.ch.empty()) {
            const Nd* b = nd.ch.data();
            const Nd* e = b + nd.ch.size();
            nl(ind + 1);
            const Nd* st = b;
            for (const Nd* p = b; p <= e; ++p) {
                bool en = p == e;
                if (!en && !(p->tk.ty == Pu && p->tk.tx == "," && !p->gp)) continue;
                if (st < p) rr(st, p, ind + 1, false, false);
                if (!en) {
                    ap(p->tk);
                    nl(ind + 1);
                }
                st = p + 1;
            }
            nl(ind);
        }
        ap(nd.cl);
    }

    void ln(const Nd& nd, int ind) {
        if (nd.op == 't') {
            pt(nd.tk);
            rr(nd.ch.data(), nd.ch.data() + nd.ch.size(), ind, false, false);
            ap(nd.cl);
            return;
        }
        if (nd.op == '{' && !nd.ob) {
            blk(nd, ind, false);
            return;
        }
        pt(nd.tk);
        rr(nd.ch.data(), nd.ch.data() + nd.ch.size(), ind, false, false);
        ap(nd.cl);
    }

    void blk(const Nd& nd, int ind, bool sw) {
        if (fo) {
            bd = true;
            return;
        }
        pt(nd.tk);
        if (nd.ch.empty()) {
            ap(nd.cl);
            return;
        }
        nl(ind + 1);
        rr(nd.ch.data(), nd.ch.data() + nd.ch.size(), ind + 1, true, sw);
        nl(ind);
        ap(nd.cl);
    }

    void rr(const Nd* b, const Nd* e, int ind, bool st, bool sw) {
        int tn = 0;
        int xt = 0;
        bool ss = st;
        bool lp = false;
        bool nb = false;
        for (const Nd* p = b; p != e && !bd; ++p) {
            const Nd& nd = *p;
            const Nd* pv = p != b ? p - 1 : nullptr;
            const Nd* nx = p + 1 != e ? p + 1 : nullptr;
            const Tk& tk = nd.tk;
            if (st && ss) {
                bool dc = !nd.gp && ((kn(tk, "function") || kn(tk, "class")) ||
                                     (tk.ty == Id && tk.tx == "async" && nx && !nx->gp && kn(nx->tk, "function")));
                bool bg = false;
                if (pv && !fo && tk.ln > le(lt(*pv)) + 1) bg = true;
                if (ind == 0 && pv && (dl || dc)) bg = true;
                if (ind == 0) dl = dc;
                if (bg) {
                    bk();
                }
                if (sw && !nd.gp && (kn(tk, "case") || kn(tk, "default"))) {
                    lp = true;
                    xt = 0;
                }
                nl(ind + xt);
                ss = false;
            } else if (st && pv && tk.nl && nq(lt(*pv), tk)) {
                nl(ind + xt);
            }
            if (nd.gp) {
                bool nbk = false;
                if (nd.op == '(' && pv && !pv->gp) {
                    nbk = (pv->tk.ty == Kw && hk.count(pv->tk.tx)) || (pv->tk.ty == Id && pv->tk.tx == "go2jsSetSourceLines");
                }
                if (nd.op == '{' && !nd.ob) {
                    bool isw = pv && pv->gp && pv->op == '(' && p - 1 != b && !(p - 2)->gp && kn((p - 2)->tk, "switch");
                    blk(nd, ind, isw);
                    if (st && !fo) {
                        bool sm = false;
                        if (nx) {
                            if (nx->gp) {
                                sm = nx->op == '(' || nx->op == '[';
                            } else {
                                const Tk& nt = nx->tk;
                                sm = (nt.ty == Kw && (nt.tx == "else" || nt.tx == "catch" || nt.tx == "finally")) ||
                                     (nt.ty == Pu && (nt.tx == ")" || nt.tx == "," || nt.tx == ";" || nt.tx == "." || nt.tx == "?." || nt.tx == "]")) ||
                                     (kn(nt, "while") && pv && !pv->gp && kn(pv->tk, "do"));
                            }
                        }
                        if (nx && !sm) {
                            nl(ind + xt);
                            ss = true;
                        }
                    }
                } else {
                    ex(nd, ind, nbk);
                }
                continue;
            }
            if (tk.ty == Hb) {
                pt(tk);
                nl(ind);
                ss = st;
                continue;
            }
            if (tk.ty == Pu && tk.tx == ";") {
                pt(tk);
                if (st) {
                    nl(ind + xt);
                    ss = true;
                }
                continue;
            }
            if (tk.ty == Pu && tk.tx == "?") {
                ++tn;
                pt(tk);
                continue;
            }
            if (tk.ty == Pu && tk.tx == ":") {
                kc = tn == 0;
                if (!kc) --tn;
                pt(tk);
                if (lp && kc) {
                    lp = false;
                    xt = 1;
                    nl(ind + xt);
                    ss = st;
                }
                kc = false;
                continue;
            }
            pt(tk);
            (void)nb;
        }
    }
};

}

std::string pf(const Ts& ts) {
    Nd rt;
    size_t ix = 0;
    while (ix < ts.size()) ps(ts, ix, rt);
    Fm fm;
    fm.rr(rt.ch.data(), rt.ch.data() + rt.ch.size(), 0, true, false);
    fm.nl(0);

    std::vector<size_t> pl;
    size_t ph = 1;
    for (const std::string& ln : fm.ls) {
        pl.push_back(ph);
        ph += 1 + static_cast<size_t>(std::count(ln.begin(), ln.end(), '\n'));
    }

    for (size_t tl = 0; tl < fm.ls.size(); ++tl) {
        std::string& tx = fm.ls[tl];
        size_t ix = tx.find("go2jsSetSourceLines({");
        if (ix == std::string::npos || tx.find_first_not_of(' ') != ix) continue;

        std::unordered_map<size_t, std::string> ol;
        size_t ps = ix + 21;
        while (ps < tx.size() && tx[ps] != '}') {
            size_t ky = 0;
            while (ps < tx.size() && std::isdigit(static_cast<unsigned char>(tx[ps]))) ky = ky * 10 + static_cast<size_t>(tx[ps++] - '0');
            ps += 2;
            size_t qe = tx.find('"', ps + 1);
            if (qe == std::string::npos) break;
            ol[ky] = tx.substr(ps, qe - ps + 1);
            ps = qe + 1;
            while (ps < tx.size() && (tx[ps] == ',' || tx[ps] == ' ')) ++ps;
        }

        size_t lo = fm.ol[tl];
        std::string nw = tx.substr(0, ix) + "go2jsSetSourceLines({";
        bool fs = true;
        for (size_t kx = tl + 1; kx < fm.ls.size(); ++kx) {
            if (fm.ol[kx] <= lo) continue;
            auto it = ol.find(fm.ol[kx] - lo);
            if (it == ol.end()) continue;
            if (!fs) nw += ", ";
            fs = false;
            nw += std::to_string(pl[kx] - pl[tl]) + ": " + it->second;
        }
        nw += "});";
        tx = nw;
    }

    std::string ot;
    for (const std::string& ln : fm.ls) {
        ot += ln;
        ot += '\n';
    }
    return ot;
}

}
