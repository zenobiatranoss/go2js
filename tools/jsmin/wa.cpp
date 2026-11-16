// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <cstdlib>
#include <cstring>

namespace {

std::string ot;

}

extern "C" {

__attribute__((export_name("ja")))
char* ja(unsigned long ln) {
    return static_cast<char*>(std::malloc(ln ? ln : 1));
}

__attribute__((export_name("jf")))
void jf(char* pt) {
    std::free(pt);
}

__attribute__((export_name("jr")))
const char* jr(const char* sr, unsigned long ln, int nn) {
    std::string in(sr, ln), er;
    jm::Ts ts = jm::lx(in, &er);
    if (!er.empty()) {
        ot = er;
        return nullptr;
    }
    if (!nn) jm::rn(ts);
    ot = jm::em(ts);
    if (!jm::ok(ts, jm::lx(ot))) {
        ot = "the result does not read back as the same program";
        return nullptr;
    }
    return ot.c_str();
}

__attribute__((export_name("jv")))
const char* jv(const char* sr, unsigned long ln) {
    std::string in(sr, ln), er;
    jm::Ts ts = jm::lx(in, &er);
    if (!er.empty()) {
        ot = er;
        return nullptr;
    }
    jm::Vw vv = jm::vw(ts);
    if (!vv.fd) {
        ot = "no program found";
        return nullptr;
    }
    jm::Ts pg = jm::sc(ts, vv.bi, vv.ei);
    jm::pc(pg);
    ot = jm::pf(pg);
    return ot.c_str();
}

__attribute__((export_name("jl")))
unsigned long jl() {
    return ot.size();
}

}
