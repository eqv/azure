package cpu

import (
	"errors"
	"strconv"
	"fmt"
	"os"
	"io"
	"bufio"
	"encoding/binary"
	"bytes"
	"log"
	"time"
	"math/rand" 
)

type CPU struct {
	Registers map[Register]int
	Memory    []int
	InstrCache map[int]*Instruction
	Age int
}

func NewCPU(ip int, mem []int, instr_cache map[ing]*Instruction) (cpu *CPU) {
	cpu := new(CPU)
	InitCPU(cpu, ip, mem, instr_cache)
	return cpu
}

func InitCPU(cpu *CPU, start_ip int, mem []int, instr_cache map[ing]*Instruction) {
	cpu.Registers = make(map[Register]int)
	cpu.InstrCache = instr_cache
	cpu.mem = mem
	cpu.Reinit()
	cpu.Registers[ip] = start_ip
	return
}

func (cpu *CPU) reinit(){
	cpu.Age = 0
	cpu.Registers[eq] = 0
	cpu.Registers[smaller] = 0
	cpu.Registers[bigger] = 0
	cpu.Registers[t0] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t1] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t2] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t3] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t4] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t5] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t6] = rand.Int31n(len(cpu.mem))
	cpu.Registers[t7] = rand.Int31n(len(cpu.mem))
	cpu.Registers[ip] = rand.Int31n(len(cpu.mem))
}

func dec(val int) string {
	return strconv.FormatInt(int64(val), 10)
}

func hex(val int) string {
	return "0x"+strconv.FormatInt(int64(val), 16)
}

func (c *CPU) InspectIP() (res string) {
	ipv := c.Registers[ip]
	res = "("+hex(ipv)+") "
	return res
}

func (c *CPU) Inspect() (res string) {
	res =""
	res += "ip: "+c.InspectIP()+"  , "
	res += "e: "+hex(c.Registers[eq])+", "
	res += "s: "+hex(c.Registers[smaller])+", "
	res += "b: "+hex(c.Registers[bigger])+", "
	res += "t0: "+hex(c.Registers[t0])+", "
	res += "t1: "+hex(c.Registers[t1])+", "
	res += "t2: "+hex(c.Registers[t2])+", "
	res += "t3: "+hex(c.Registers[t3])+", "
	res += "t4: "+hex(c.Registers[t4])+", "
	res += "t5: "+hex(c.Registers[t5])+", "
	res += "t6: "+hex(c.Registers[t6])+", "
	res += "t7: "+hex(c.Registers[t7])+"\n"
	return res
}

func (c *CPU) Decompile(val int) (*Instruction,error) {
	instr,err := Dissect(val,c)
	if err != nil {
		return nil, errors.New("Undable to Decode Instruction: "+strconv.FormatInt(int64(val),16)+" reason: "+err.Error())
	}
	return instr,nil
}

func (c *CPU) GetInstr(instrval int) (*Instruction,error) {
	if val,ok := c.InstrCache[instrval]; ok {
		return val,nil
	}
	instr, err := c.Decompile(instrval)
	if err!=nil {return nil,err}
	c.InstrCache[instrval]=instr
	return instr,nil
}

func (c *CPU) Log(v ...interface{}) {
	log.Print("CPU [",0,"] ",v)
}

func (c *CPU) Tick() error{
	c.Age+=1
	if c.Age > 10000 && rand.Int31n(1000)==0 {return errors.New("died of old age")}
	ipval,err := c.GetRegister(ip)
	if err != nil {return errors.New("unable to get ip") }
	instrval, err := c.GetMemory(ipval)
	if err != nil {return errors.New("unable to get cmd") }
	instr, err := c.GetInstr(instrval)
	if err != nil {return errors.New("unable to decode instr") }
	//fmt.Println(instr.Inspect())
	return instr.Exec(c)
}

func (c *CPU) GetRegister(reg Register) (int, error) {
	if val, ok := c.Registers[reg]; ok {
		return val, nil
	}
	return 0, errors.New("Invalid Register access (read): "+ hex(int(reg)))
}

func (c *CPU) SetRegister(reg Register, val int) error {
	if _, ok := c.Registers[reg]; ok {
		c.Registers[reg] = val
		return nil
	}
	return errors.New("Invalid Register access (write)"+ hex(int(reg)))
}

func (c *CPU) GetMemory(addr int) (int, error) {
	if addr >= len(c.Memory) || addr < 0 {
		return 0, errors.New("Invalid Memory access (read)")
	}
	return c.Memory[addr], nil
}

func (c *CPU) SetMemory(addr int, value int) error {
	if addr > len(c.Memory) || addr < 0 {
		return errors.New("Invalid Memory access (write)")
	}
	c.Memory[addr] = value
	return nil
}

func (c *CPU) Sys(num Syscall, arg1 int) error {
	arg2, err := c.GetRegister(t0)
	if err != nil {return err}
	switch num {
		case EXIT : c.Terminated = true
	}
	return nil
}
