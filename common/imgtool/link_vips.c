// link_vips.c
extern void vips_foreign_load_webp_init (void);
extern void vips_foreign_load_heif_init (void);

__attribute__((constructor)) static void force_link_vips_formats(void) {
    vips_foreign_load_webp_init();
    vips_foreign_load_heif_init();
}
