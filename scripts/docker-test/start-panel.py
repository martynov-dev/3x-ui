"""Run the test panel with the service unit's umask and capability ceiling."""
import ctypes
import os
from pathlib import Path

unit = dict(
    line.split("=", 1)
    for line in Path("/test/x-ui.service.debian").read_text().splitlines()
    if line.startswith(("UMask=", "CapabilityBoundingSet="))
)
capabilities = {
    "CAP_CHOWN": 0,
    "CAP_DAC_OVERRIDE": 1,
    "CAP_NET_BIND_SERVICE": 10,
    "CAP_NET_ADMIN": 12,
    "CAP_NET_RAW": 13,
}
allowed = {capabilities[name] for name in unit["CapabilityBoundingSet"].split()}
libc = ctypes.CDLL(None, use_errno=True)
for capability in range(int(Path("/proc/sys/kernel/cap_last_cap").read_text()) + 1):
    if capability not in allowed and libc.prctl(24, capability, 0, 0, 0) != 0:
        raise OSError(ctypes.get_errno(), "Cannot restrict panel capabilities")
if libc.prctl(38, 1, 0, 0, 0) != 0:
    raise OSError(ctypes.get_errno(), "Cannot set no_new_privs")
os.umask(int(unit["UMask"], 8))
os.execv("/app/x-ui", ["/app/x-ui"])
