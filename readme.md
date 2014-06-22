#README

This is a small custom VM written in golang that was used as a service during the 2012 rwthctf. 

## Structure
The code of the implements the VM itself is stored in `./cpu/`. The original service image is stored in `./data/`. 
A simple makro assembler containing complex makros suchas function definitions/calls and controllflow is given in `compiler.rb`. A set of primities for strings (strcopy println printi itou strlen etc), heap (malloc/free) and some crypto (weak hash/stream cipher) are implemented in `string.rb`, `libcrypt.rb`, `libmem.rb` and `string.rb`. Finally the code of the service is implemented in `service.rb`.

The following document contains the information given to the participants.

This sevice is an image for a VM simulating a custom CPU. For every incoming connection,
one such VM is spawned and stdin/stdout fds (0 and 1) are mapped to the corresponding socket.
The challenge is not to understand the vm (even though you may want to take a few looks in the code to understand what exactly certain instructions do). To the best of my knowledge there should be no exploitable bugs in the VM itself (go newb though). It is my intend to have an instruction set that is simple enough to write some interesting analysis tools during the CTF (you will have a hard time reading the code without). Thus the VM has a very small set of instructions. On the other hand this makes an macro assembler necessary for implementing programs. You can find a copy of the VM code, the assembler and even some parts of the code of the service in /home/service_sources. You have to recompile the VM with make.sh if you change the bytecode in data/img.go. Recompiling the VM takes the image from data/img.go and the rest of the code to produce a standalone binary. For make.sh to be successfull the service has to be currently stoped, because the original binary is locked from the running process otherwise.

```sh
root@vuln $ sv stop azurecoast
root@vuln $ pkill binary
root@vuln $ su service_source
serv~@vul $ cd service/go/src/rwthctf && sh make.sh
serv~@vul $ exit
root@vuln $ sv start azurecoast
```

I also added the assembler I used to generate the image (compiler.rb). Once you obtained a proper disassembly of the service, you can use compiler.rb to recompile it.
WARNING: if you run compiler.rb two things will happen: 1) it will crash (unless you have your own disassembled version of the service), because parts of the program are missing in the asm version and 2) it may overwrite data/img.go so don't do this unless you know what you are doing. Have fun reversing the byte code array in data/img.go :)

##compiler.rb

compiler.rb is a powerfull makro assembler thar runs under ruby1.9.3. Besides the primitiv instructions it supports a wide range of makros such as call, push/pop, function definitions, labels, if_then, if_else etc.
To use compiler.rb you will have to write your code into a CodeGen object. This works by creating such an object
and calling the asm mehtod with a block containing you code.

```ruby
  code = CodeGen.new
  code.asm do
    mov t1,4
    inc [t1]
  end
```

you can also "include" classes that have a method code(gen) which calls gen.asm in the same manner (see libmem.rb). This is especially usefull for you since you can output your code into a skeleton file containting just:

```ruby
  class Disassembly
    def code(asm)
      asm.asm do
        %{disassembly}
      end
    end
  end
```

and then replace the code in compiler.rb

```ruby
  asm = CodeGen.new
  asm.asm do
    ldw t0, ref(:entermain)
    jmp_to :init_mem
    ...
  end
```

by just (not you also have to require the file itself in the begining of compiler.rb).

```ruby
  asm.asm do
    import(Disassembly.new)
  end
```

labels are added with:

```ruby
  label :name
```

and referenced with

```ruby
  ldw t1, ref(:name)
  jmp t1
```

you can place arbitrary data with the data macro

```ruby
  data(0x12345)
```

you can push/pop multiple values from the stack with the push/pop macros

```ruby
  push t1,t2,t3
  pop t1,t2,t3  #note that the order for pop is inverted so that this will NOT change any registers
```

you can call to a arbitrary address by using the `call(target)` macro. The `call_to(:label)` macro will also handle getting the `ref(:label)` for you (same goes for the `jmp_to` macro). There are more macros (`get,set`, `if_then`, `if_else` etc.) but I'm to lazy to document them all - have a look in the supplied code( libmem.rb etc) if you want to use / understand / replace them.

##CPU specs

###WORDSIZE
The smallest addressable unit is a 4 BYTE word. Every instruction is 4 byte
long.  The only exception to this is the `ldw` instruction which uses two machine
words (the first one is the instruction, the second one is the word that is
    loaded into the dst of the instruction.

###REGISTER
The CPU has the following registers: `ip, eq, smaller, bigger, t0, ..., t7`
* `ip` is the program counter
* `eq`, `smaller` and `bigger` are set to 0/1 depending on the result of the last
arithmic operation (eq to 0, smaller than etc).
* `t0` to `t7` are general purpose registers. By convention `t7` is used as a stack pointer, `t6` is used in macros and should not be used. `t0` and `t1` are registers used to supply additional arguments to syscalls

###INSTRUCTIONS
All instructions are of the kind `[op dst src]`. However in some instructions src
may be ignored.  dst and src may both be either a constant in from 0 to 255, a
register or a register derefernce.

```
	examples:
  mov t0,3 #copies 3 into register t0
  mov [t0],3 #copies 3 into the memory cell at *t0
```

The cpu understands the following operations:
Instructions = `[:add, :sub, :mul, :div, :mod, :rol, :band, :bor, :not, :xor, :cmp, :mov, :ldw, :jmp, :jnz, :jz, :sys]`

* `add`, `sub`, `mul`, `div`, `mod`, `rol`, `band`, `bor`, `not`, `xor` should be self explainatory
  (all of them set `eq`,`bigger`,`smaller* according to the result compared to 0)
* `cmp` sets eq bigger and smaller according to dst compared to src
* `mov` copies src to dest
* `ldw` will only use the dst field and the next word in memory and copy the
  content of the next word into dst
    example (this will load 0x12356 into t1)

      ldw t1,0
      0x12356

* `jmp` jump to dst
* `jnz` jz will jump to dst if src is != 0 or == 0 respectively
* `sys` performs a syscall. The index is stored in dst, the first argument in src, the second argument in t0. `sys` may change t0 and t1 to return values

###ENCODING
See cpu/dissassembler.go or compiler.rb if you want to dissassemble on word into one instruction

###SYSCALLS
There are a few syscalls:

* EXIT = 0 #terminates the vm, arg1,arg2 unused, does not return
* READB = 1 #reads one byte from fd arg1 and stores it in t0, returns 1 in t1 if read was successfull, 0 else
* WRITEB = 2 #writes one byte from arg2 to fd arg1, returns nothing
* READW = 3 #reads one word from fd arg1 and stores it in t0, returns 1 in t1 if read was successfull, 0 else
* WRITEW = 4 #writes arg2 to fd arg1, returns nothing
* EXEC = 5 #reads the string arg1 points to from the VM memory and executes it as shell instruction, returns a fd for the stdin/stdout pipe of the process in t0, returns 0 if starting failes
* BREAK = 6 #stops the CPU until enter is pressed on the stdin of the service (should not be used in production code)
* STEP = 7 #sets the single step flag to arg1 (1 = singlesteping, 0 = stop singlestepping). While stepping the cpu state is printed to stdout of the service and after every instruction enter has to be pressed
* OPEN = 8 #opens the file with path given as "./storage/"+get_string_from_VM_memory(arg1) rw, returning the fd in t0 (0 if opening failed)
* CORE = 9 #returns some informations about the core in t0 (size of memory) and t1 (size of initial code segment)
* CLOSE = 10 #closes the fd given by arg1
* CLOCK = 11 #sets t0 to the current time


##GLHF
I hope this service will be fun to explore and that there is as no annoying guesswork necessary
If you code some cool stuff to analyse the image (or some polymorphic shellcode gnereator or whatever), let me know (coco@hexgolems.com). Looking forward to seeing you solutions,
happy hacking and may the force be with you.
