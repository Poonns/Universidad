package 2º.EU.Ejercicios;

public interface Geometria {
    float area();
}

public class Cilindro implements Geometria {
    private float radio;
    private float altura;
    private double pi;


    public Cilindro(float radio, float altura, double pi) {
        this.radio = radio;
        this.altura = altura;
        this.pi = 3.1416;
    }


    public float area() {
        return 2 * pi * radio * (altura + radio);
    }
}


    public class Rectangulo implements Geometria {
        private float base;
        private float altura;
    
        public Rectangulo(float base, float altura) {
            this.base = base;
            this.altura = altura;
        }
    
        public float area() {
            return base * altura;
        }
    }
    
}