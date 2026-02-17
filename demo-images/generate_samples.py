#!/usr/bin/env python3
"""
Generate sample images with variations for perceptual hashing demo.
These images will demonstrate similar photos that perceptual hashing can detect.
"""

from PIL import Image, ImageEnhance, ImageFilter
import os

def create_base_image(size=(256, 256), bg_color=(135, 206, 235), sun_color=(255, 215, 0)):
    """Create a base sunset image"""
    img = Image.new('RGB', size, bg_color)

    # Create a gradient for sunset
    for y in range(size[1]):
        progress = y / size[1]
        r = int(bg_color[0] + (255 - bg_color[0]) * progress)
        g = int(bg_color[1] + (100 - bg_color[1]) * progress)
        b = int(bg_color[2] + (50 - bg_color[2]) * progress)
        for x in range(size[0]):
            img.putpixel((x, y), (r, g, b))

    # Add sun
    sun_radius = size[0] // 8
    sun_x = size[0] // 2
    sun_y = size[1] // 3

    for y in range(size[1]):
        for x in range(size[0]):
            dist = ((x - sun_x) ** 2 + (y - sun_y) ** 2) ** 0.5
            if dist < sun_radius:
                img.putpixel((x, y), sun_color)

    return img

def create_base_cat_image(size=(256, 256), bg_color=(240, 248, 255)):
    """Create a simple cat silhouette image"""
    img = Image.new('RGB', size, bg_color)

    # Cat body (simple shapes)
    cat_color = (100, 100, 100)

    # Body (ellipse)
    for y in range(size[1]):
        for x in range(size[0]):
            # Body
            cx, cy = size[0] // 2, size[1] // 2 + 20
            if ((x - cx) ** 2) / 40 ** 2 + ((y - cy) ** 2) / 50 ** 2 <= 1:
                img.putpixel((x, y), cat_color)

            # Head
            hx, hy = size[0] // 2, size[1] // 3
            if ((x - hx) ** 2) / 35 ** 2 + ((y - hy) ** 2) / 30 ** 2 <= 1:
                img.putpixel((x, y), cat_color)

            # Left ear
            ex1, ey1 = hx - 20, hy - 15
            ex2, ey2 = hx - 10, hy - 35
            ex3, ey3 = hx - 5, hy - 15
            # Simple triangle check
            if point_in_triangle(x, y, ex1, ey1, ex2, ey2, ex3, ey3):
                img.putpixel((x, y), cat_color)

            # Right ear
            ex1, ey1 = hx + 20, hy - 15
            ex2, ey2 = hx + 10, ey2
            ex3, ey3 = hx + 5, ey3
            if point_in_triangle(x, y, ex1, ey1, ex2, ey2, ex3, ey3):
                img.putpixel((x, y), cat_color)

    return img

def point_in_triangle(px, py, x1, y1, x2, y2, x3, y3):
    """Check if point is in triangle using barycentric coordinates"""
    def area(ax, ay, bx, by, cx, cy):
        return abs((ax * (by - cy) + bx * (cy - ay) + cx * (ay - by)) / 2.0)

    A = area(x1, y1, x2, y2, x3, y3)
    A1 = area(px, py, x2, y2, x3, y3)
    A2 = area(x1, y1, px, py, x3, y3)
    A3 = area(x1, y1, x2, y2, px, py)

    return abs(A - (A1 + A2 + A3)) < 0.1

def create_variations(base_img, output_dir, prefix, count=5):
    """Create variations of base image"""
    variations = []

    for i in range(count):
        img = base_img.copy()

        # Apply different transformations based on index
        if i == 1:
            # Slightly brighter
            enhancer = ImageEnhance.Brightness(img)
            img = enhancer.enhance(1.15)
            filename = f"{prefix}_brightened.jpg"
        elif i == 2:
            # Slightly more contrast
            enhancer = ImageEnhance.Contrast(img)
            img = enhancer.enhance(1.2)
            filename = f"{prefix}_contrast.jpg"
        elif i == 3:
            # Slight blur
            img = img.filter(ImageFilter.GaussianBlur(radius=0.5))
            filename = f"{prefix}_blurred.jpg"
        elif i == 4:
            # Saturation boost (Instagram-like)
            enhancer = ImageEnhance.Color(img)
            img = enhancer.enhance(1.3)
            filename = f"{prefix}_saturated.jpg"
        else:
            # Original
            filename = f"{prefix}_original.jpg"

        filepath = os.path.join(output_dir, filename)
        img.save(filepath, 'JPEG', quality=90)
        variations.append(filepath)
        print(f"Created: {filename}")

    return variations

def main():
    output_dir = '/home/ubuntu/file-deduplicator/demo-images/samples'
    os.makedirs(output_dir, exist_ok=True)

    print("Generating sample images for perceptual hashing demo...")
    print("=" * 60)

    # Create sunset variations
    print("\n🌅 Creating sunset image variations...")
    sunset_base = create_base_image()
    sunset_files = create_variations(sunset_base, output_dir, 'sunset')

    # Create cat variations
    print("\n🐱 Creating cat image variations...")
    cat_base = create_base_cat_image()
    cat_files = create_variations(cat_base, output_dir, 'cat')

    print("\n" + "=" * 60)
    print(f"✅ Generated {len(sunset_files) + len(cat_files)} sample images")
    print(f"📁 Location: {output_dir}")
    print("\nThese images have slight variations that perceptual hashing can detect:")
    print("  - Original")
    print("  - Brightened (+15%)")
    print("  - Contrast boosted (+20%)")
    print("  - Slightly blurred")
    print("  - Saturated (+30%)")

if __name__ == '__main__':
    main()
