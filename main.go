package main
import (
		"azure/cpu"
		"azure/data"
		"time"
		"net"
		"log"
		"os"
		"runtime"
		"os/exec"
		"fmt"
		"math/rand" 
//		"strconv"
	)

func simulate(){
	coresize:=1048576
	instr_cache	:= make(map[int]*Instruction)
	memory			:= make([]int, coresize)
	cpus := make([]*cpu,256)
	for i,_ :=range cpus {
		cpus[i] = cpu.NewCPU(rand.Int31n(coresize), memory, instr_cache)
	}

	for addr, _ := range mem { //initialize core memory
		mem[addr] = rand.Int32()
	}  

	tickcount := 0
	for { //simulate core (nearly) endlessly
		for _,core :=range cpus {
		err := core.Tick()
			if err != nil { // something went wrong while ticking the core
				core.Reinit();
			}
		}
		tickcount+=1;
		if tickcount & 0xffff == 0 { //yield every few ticks so other cores can run
			runtime.Gosched()
		}
	}
}


func main () {
	simulate()
}
