package main
import (
		"azure/cpu"
		"runtime"
//		"fmt"
		"math/rand" 
//		"strconv"
	)

func simulate(){
	//mov_to_t1 := 0x0b050503
	coresize:=1048576
	instr_cache	:= make([]*(cpu.Instruction), coresize)
	memory			:= make([]int, coresize)
	cpus := make([]*cpu.CPU,256)
	for i,_ :=range cpus {
		cpus[i] = cpu.NewCPU(rand.Intn(coresize), memory, instr_cache)
	}

	for addr, _ := range memory { //initialize core memory
		memory[addr] = rand.Int()
	}

	tickcount := 0
	thread_ages := 0
	thread_dead_count := 0
	for { //simulate core endlessly
		for _,core :=range cpus {
			for i := 0; i<10; i++ {
				err := core.Tick()
				if err != nil { // something went wrong while ticking the core
					thread_dead_count += 1
					thread_ages += core.Age
					break;
				}
			}
			core.Reinit();
		}
		tickcount+=1;
		if tickcount % 10000 == 0 { //yield every few ticks so other cores can run
			print("ticks: ",tickcount,"\n")
			opcodes := make(map[string]int)
			opcodes["bad"] = 0
			for addr,val := range memory {
				instr,err := cpus[0].GetInstr(val,addr)
				if err != nil { opcodes["bad"]+=1;continue }
				op := instr.Instr.Inspect()
				if _,ok := opcodes[op]; !ok {
					opcodes[op]=0
				}
				opcodes[op]+=1
			}
			for op,count := range opcodes {
				print(op,"\t has a prevalence of ",(float64)(count)/(float64)(len(instr_cache)),"(",count,")","\n")
			}
			avg_lifespan := (float64)(thread_ages)/(float64)(thread_dead_count)
			print("avg lifespan: ", avg_lifespan, "\n")
			thread_ages = 0
			thread_dead_count = 0
			runtime.Gosched()
		}
	}
}


func main () {
	simulate()
}
