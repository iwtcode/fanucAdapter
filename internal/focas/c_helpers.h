#ifndef C_HELPERS_H
#define C_HELPERS_H

#include "fwlib32.h"

short go_cnc_startupprocess(unsigned short mode, const char* logpath);
short go_cnc_allclibhndl3(const char* ip, unsigned short port, long timeout_ms, unsigned short* handle_out);
short go_cnc_freelibhndl(unsigned short h);
short go_cnc_exeprgname(unsigned short h, char* name_out, int name_cap, long* onum_out);
short go_cnc_statinfo(unsigned short h, ODBST* stat_out);
short go_cnc_sysinfo(unsigned short h, ODBSYS* sys_info_out);
short go_cnc_rdaxisname(unsigned short h, short axis, ODBAXISNAME* axisname_out);
short go_cnc_absolute(unsigned short h, short axis, short length, ODBAXIS* axis_out);
short go_cnc_relative(unsigned short h, short axis, short length, ODBAXIS* axis_out);
short go_cnc_machine(unsigned short h, short axis, short length, ODBAXIS* axis_out);
short go_cnc_rdposition(unsigned short h, short type, short* data_num, ODBPOS* position);

// Новые функции для чтения программы
short go_cnc_upstart(unsigned short h, short prog_num);
short go_cnc_upload(unsigned short h, ODBUP* data_out, unsigned short* len);
short go_cnc_upend(unsigned short h);
short go_cnc_rdexecprog(unsigned short h, unsigned short* length, short* blknum, char* data);

#endif // C_HELPERS_H