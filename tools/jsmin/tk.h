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
};

using Ts = std::vector<Tk>;

Ts lx(const std::string& sr);
std::string em(const Ts& ts);
size_t rn(Ts& ts);
bool rw(const std::string& wd);
bool ok(const Ts& a, const Ts& b);

}
