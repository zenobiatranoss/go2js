CXX ?= g++
CXXFLAGS ?= -O2 -std=c++17 -Wall -Wextra
OB = cl.o lx.o em.o rn.o

go2js-min: $(OB)
	$(CXX) $(CXXFLAGS) -o $@ $(OB)

%.o: %.cpp tk.h
	$(CXX) $(CXXFLAGS) -c -o $@ $<

clean:
	rm -f $(OB) go2js-min
