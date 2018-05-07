# include the top level Makefile.global to ensure that we can read the configuration variable
top_builddir = ../..
include $(top_builddir)/src/Makefile.global

ifeq ($(enable_go_utils),yes)
all:
	$(MAKE) -C src/gp_upgrade $@

clean distclean:
	$(MAKE) -C src/gp_upgrade clean

install:
# allow someone in src/gp_upgrade to run `make install` without dependencies
# but include it here for the build system
	$(MAKE) -C src/gp_upgrade dependencies $@

# do nothing if global configutils is not enabled
else
all:
	@echo "Configuration not enabled for go-utils."

endif
