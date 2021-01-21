#include <stdlib.h>

void Run();

__attribute__((constructor)) static void init(int argc, char **argv, char **envp)
{
    unsetenv("DYLD_INSERT_LIBRARIES");
    Run();
}