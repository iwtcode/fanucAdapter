package focas

/*
#cgo CFLAGS: -I../../
#cgo LDFLAGS: -L../../ -lfwlib32 -Wl,-rpath,'$ORIGIN'
// #cgo windows LDFLAGS: -L../../ -lfwlib32

#include <string.h>
#include "c_helpers.h"

short go_cnc_startupprocess(unsigned short mode, const char* logpath) {
    return cnc_startupprocess(mode, logpath);
}

short go_cnc_allclibhndl3(const char* ip, unsigned short port, long timeout_ms, unsigned short* handle_out) {
    return cnc_allclibhndl3(ip, port, timeout_ms, handle_out);
}

short go_cnc_freelibhndl(unsigned short h) {
    return cnc_freelibhndl(h);
}

short go_cnc_exeprgname(unsigned short h, char* name_out, int name_cap, long* onum_out) {
    ODBEXEPRG p;
    short rc = cnc_exeprgname(h, &p);
    if (rc == EW_OK) {
        int n = (int)sizeof(p.name);
        if (n >= name_cap) n = name_cap - 1;
        memcpy(name_out, p.name, n);
        name_out[n] = '\0';
        *onum_out = p.o_num;
    }
    return rc;
}

short go_cnc_statinfo(unsigned short h, ODBST* stat_out) {
	return cnc_statinfo(h, stat_out);
}

short go_cnc_sysinfo(unsigned short h, ODBSYS* sys_info_out) {
    return cnc_sysinfo(h, sys_info_out);
}

short go_cnc_rdaxisname(unsigned short h, short axis, ODBAXISNAME* axisname_out) {
    return cnc_rdaxisname(h, &axis, axisname_out);
}

short go_cnc_absolute(unsigned short h, short axis, short length, ODBAXIS* axis_out) {
    return cnc_absolute(h, axis, length, axis_out);
}

short go_cnc_relative(unsigned short h, short axis, short length, ODBAXIS* axis_out) {
    return cnc_relative(h, axis, length, axis_out);
}

short go_cnc_machine(unsigned short h, short axis, short length, ODBAXIS* axis_out) {
    return cnc_machine(h, axis, length, axis_out);
}

short go_cnc_rdposition(unsigned short h, short type, short* data_num, ODBPOS* position) {
    return cnc_rdposition(h, type, data_num, position);
}
*/
import "C"
