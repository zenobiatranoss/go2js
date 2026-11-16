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
    std::fputs("usage: go2js-min [-o out] [-n] [-k] [-s] [-p|-u|-S [-e]] [file|-]\n"
               "  -o  write to a file instead of standard output\n"
               "  -n  do not shorten the names of the runtime\n"
               "  -k  lex the result again and fail if it differs from what was meant\n"
               "  -s  report the sizes on standard error\n"
               "  -p  write the script readable, indented and in lines, instead of small\n"
               "  -u  write only the program of a generated script, readable\n"
               "  -S  write the program to -o and the runtime beside it as NAME.runtime.js\n"
               "  -e  with -S, join the two files with import and export instead of require\n",
               stderr);
    return 2;
}

std::string bs(const std::string& pa) {
    size_t sl = pa.find_last_of('/');
    std::string fl = sl == std::string::npos ? pa : pa.substr(sl + 1);
    size_t dt = fl.rfind('.');
    return dt == std::string::npos || dt == 0 ? fl : fl.substr(0, dt);
}

std::string dr(const std::string& pa) {
    size_t sl = pa.find_last_of('/');
    return sl == std::string::npos ? std::string() : pa.substr(0, sl + 1);
}

}

int main(int ac, char** av) {
    std::string in = "-", ou;
    bool nn = false, kk = false, ss = false, pp = false, uu = false, sp = false, es = false;

    for (int ix = 1; ix < ac; ++ix) {
        if (!std::strcmp(av[ix], "-o") && ix + 1 < ac) {
            ou = av[++ix];
        } else if (!std::strcmp(av[ix], "-n")) {
            nn = true;
        } else if (!std::strcmp(av[ix], "-k")) {
            kk = true;
        } else if (!std::strcmp(av[ix], "-s")) {
            ss = true;
        } else if (!std::strcmp(av[ix], "-p")) {
            pp = true;
        } else if (!std::strcmp(av[ix], "-u")) {
            uu = true;
        } else if (!std::strcmp(av[ix], "-S")) {
            sp = true;
        } else if (!std::strcmp(av[ix], "-e")) {
            es = true;
        } else if (av[ix][0] == '-' && av[ix][1] != '\0') {
            return us();
        } else {
            in = av[ix];
        }
    }

    try {
        std::string sr = rd(in);
        std::string er;
        jm::Ts ts = jm::lx(sr, &er);
        if (!er.empty()) throw std::runtime_error(er);
        if (sp) {
            if (ou.empty() || ou == "-") throw std::runtime_error("-S needs -o");
            std::string rt, pr, ms;
            if (!jm::sx(sr, ts, bs(ou), es, rt, pr, ms)) throw std::runtime_error(ms);
            if (kk) {
                std::string e1, e2;
                jm::lx(rt, &e1);
                jm::lx(pr, &e2);
                if (!e1.empty() || !e2.empty()) throw std::runtime_error("a file of the split does not read back: " + e1 + e2);
            }
            wr(ou, pr);
            wr(dr(ou) + bs(ou) + ".runtime.js", rt);
            if (ss) std::fprintf(stderr, "program %zu bytes, runtime %zu bytes\n", pr.size(), rt.size());
            return 0;
        }

        if (pp || uu) {
            jm::Ts pg = ts;
            if (uu) {
                jm::Vw vv = jm::vw(ts);
                if (!vv.fd) throw std::runtime_error("no program found: go2jsSetPackageName is never called");
                pg = jm::sc(ts, vv.bi, vv.ei);
            }
            jm::pc(pg);
            std::string pt = jm::pf(pg);
            if (kk && !jm::ak(pg, jm::lx(pt))) {
                std::fputs("go2js-min: the result does not read back as the same program\n", stderr);
                return 3;
            }
            wr(ou, pt);
            if (ss) std::fprintf(stderr, "%zu -> %zu bytes\n", sr.size(), pt.size());
            return 0;
        }

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
