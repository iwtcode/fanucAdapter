#ifndef C_HELPERS_H
#define C_HELPERS_H

#include "fwlib32.h"

short go_cnc_startupprocess(unsigned short mode, const char* logpath);
short go_cnc_allclibhndl3(const char* ip, unsigned short port, long timeout_ms, unsigned short* handle_out);
short go_cnc_freelibhndl(unsigned short h);
short go_cnc_exeprgname(unsigned short h, char* name_out, int name_cap, long* onum_out);
short go_cnc_statinfo(unsigned short h, ODBST* stat_out);
short go_cnc_sysinfo(unsigned short h, ODBSYS* sys_info_out);

#endif // C_HELPERS_H