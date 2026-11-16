// https://github.com/q2hp0 | sora <q2hp09@gmail.com>
#include "tk.h"

#include <cstdio>
#include <cstring>
#include <fstream>
#include <iostream>
#include <iterator>
#include <sstream>
#include <stdexcept>

namespace {

std::string rd(const std::string& pa) {
    if (pa == "-") {
        std::ostringstream os;
        os << std::cin.rdbuf();
        return os.str();
    }
    std::ifstream fl(pa, std::ios::binary);
    if (!fl) throw std::runtime_error("cannot read " + pa);
    return std::string(std::istreambuf_iterator<char>(fl), std::istreambuf_iterator<char>());
}

void wr(const std::string& pa, const std::string& dt) {
    if (pa.empty() || pa == "-") {
        std::cout.write(dt.data(), static_cast<std::streamsize>(dt.size()));
        return;
    }
    std::ofstream fl(pa, std::ios::binary);
    if (!fl) throw std::runtime_error("cannot write " + pa);
    fl.write(dt.data(), static_cast<std::streamsize>(dt.size()));
}

int us() {
    std::fputs("usage: go2js-min [-o out] [-n] [-k] [-s] [file|-]\n"
               "  -o  write to a file instead of standard output\n"
               "  -n  do not shorten the names of the runtime\n"
               "  -k  lex the result again and fail if it differs from what was meant\n"
               "  -s  report the sizes on standard error\n",
               stderr);
    return 2;
}

}

int main(int ac, char** av) {
    std::string in = "-", ou;
    bool nn = false, kk = false, ss = false;

    for (int ix = 1; ix < ac; ++ix) {
        if (!std::strcmp(av[ix], "-o") && ix + 1 < ac) {
            ou = av[++ix];
        } else if (!std::strcmp(av[ix], "-n")) {
            nn = true;
        } else if (!std::strcmp(av[ix], "-k")) {
            kk = true;
        } else if (!std::strcmp(av[ix], "-s")) {
            ss = true;
        } else if (av[ix][0] == '-' && av[ix][1] != '\0') {
            return us();
        } else {
            in = av[ix];
        }
    }

    try {
        std::string sr = rd(in);
        jm::Ts ts = jm::lx(sr);
        size_t rc = nn ? 0 : jm::rn(ts);
        std::string ot = jm::em(ts);

        if (kk && !jm::ok(ts, jm::lx(ot))) {
            std::fputs("go2js-min: the result does not read back as the same program\n", stderr);
            return 3;
        }

        wr(ou, ot);

        if (ss) {
            double pc = sr.empty() ? 0.0 : 100.0 * (1.0 - static_cast<double>(ot.size()) / static_cast<double>(sr.size()));
            std::fprintf(stderr, "%zu -> %zu bytes (%.1f%% smaller), %zu names shortened\n", sr.size(), ot.size(), pc, rc);
        }
    } catch (const std::exception& ex) {
        std::fprintf(stderr, "go2js-min: %s\n", ex.what());
        return 1;
    }
    return 0;
}
