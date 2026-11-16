// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#pragma once

#include <string>
#include <vector>

namespace jm {

enum Ty : unsigned char { Id, Kw, Nu, St, Pu, Rx, Th, Tm, Tt, Tf, Hb };

struct Tk {
    Ty ty;
    std::string tx;
    bool nl;
    size_t ln = 0;
    size_t of = 0;
};

using Ts = std::vector<Tk>;

Ts lx(const std::string& sr, std::string* er = nullptr);
std::string em(const Ts& ts);
size_t rn(Ts& ts);
bool rw(const std::string& wd);
bool ok(const Ts& a, const Ts& b);
bool ak(const Ts& a, const Ts& b);
bool nq(const Tk& pv, const Tk& cu);

std::string pf(const Ts& ts);
void pc(Ts& ts);

struct Vw {
    bool fd = false;
    size_t bi = 0;
    size_t ei = 0;
    size_t dp = 0;
};

Vw vw(const Ts& ts);
Ts sc(const Ts& ts, size_t bi, size_t ei);
bool sx(const std::string& sr, const Ts& ts, const std::string& bs, bool es, std::string& rt, std::string& pr, std::string& er);

}
