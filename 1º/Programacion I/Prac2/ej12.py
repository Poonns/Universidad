kHw = float(input('Introduce la cantidad de kHw: '))

if 0 < kHw <= 100:
    kHw = kHw * 0.05
elif 100 < kHw < 250:
    kHw = kHw - 100
    kHw = (kHw * 0.07) + 5
elif kHw > 250:
    kHw = kHw - 250
    kHw = (kHw * 0.1) + 15.5

print(f'El importe será de {kHw} euros')

