package Hector.Prac2;

import java.util.Calendar;

public class Ejercicio6 {
    public class ExcepcionFechaNoValida extends RuntimeException{
        public ExcepcionFechaNoValida(String mensaje){
            super(mensaje); //Llama al constructor de RuntimeException con un mensaje
        }
    }

    public class Fecha{

        public static final int PRIMER_AÑO = 1800;
        public static final int ÚLTIMO_AÑO = 2500;

        private int dia;
        private int mes;
        private int año;

        public Fecha(int dia, int mes, int año){

            if(año < PRIMER_AÑO || año > ÚLTIMO_AÑO){
                throw new ExcepcionFechaNoValida("El año se debe encontrar dentro del rango");
            }
            if(mes < 1 || mes > 12){
                throw new ExcepcionFechaNoValida("El año se debe encontrar dentro del rango");
            }
            if(dia < 1 || dia > diasMes(mes, año)){
                throw new ExcepcionFechaNoValida("El año se debe encontrar dentro del rango");
            }

            this.dia = dia;
            this.mes = mes;
            this.año = año;

        }

        public String toString(){
            return dia + "/" + mes + "/" + año;
        }

        public boolean equals(Object otroObjeto){

            if(this == otroObjeto){
                return true;
            }

            if(otroObjeto instanceof Fecha){
                return false;
            }

            Fecha otraFecha = (Fecha) otroObjeto;
            return dia == otraFecha.dia && mes == otraFecha.mes && año == otraFecha.año;

        }

        public int compareTo(Fecha otraFecha){
            if(this.año < otraFecha.año){
                return -1;
            } else if(this.año > otraFecha.año){
                return 1;
            }

            if(this.mes < otraFecha.mes){
                return -1;
            } else if(this.mes > otraFecha.mes){
                return 1;
            }

            if(this.dia < otraFecha.dia){
                return -1;
            } else if(this.dia > otraFecha.dia){
                return 1;
            }

            return 0;
        }

        public int getDia(){
            return this.dia;
        }

        public int getMes(){
            return this.mes;
        }

        public int getAño(){
            return this.año;
        }

        public Fecha diaSiguiente(){
            int diaSiguiente = this.dia + 1;
            int mes = this.mes;
            int año = this.año;

            if(diaSiguiente > diasMes(mes, año)){
                diaSiguiente = 1;
                mes++;
            }

            if(mes > 12){
                mes = 1;
                año++;
            }

            return new Fecha(diaSiguiente, mes, año);
        }

    }

    public static boolean esAñoBisiesto(int año){

        if(año % 4 != 0){
            return false;
        }

        if(año % 100 == 0 && año % 400 != 0){
            return false;
        }

        return true;
    }

    public static int diasMes(int mes, int año){
        if(mes == 1  || mes == 3  || mes == 5  || mes == 7  || mes == 8  || mes == 10  || mes == 12){
            return 31;
        } else if(mes == 4  || mes == 6  || mes == 9  || mes == 11){
            return 30;
        } else if(mes == 2){
            if(esAñoBisiesto(año)){
                return 29;
            } else{
                return 28;
            }
        } else{
            throw new ExcepcionFechaNoValida("Mes no valido. Debe de ir del 1 al 12");
        }
    }

    public static Fecha hoy(){
        Calendar calendario = Calendar.getInstance();
        int dia = calendario.get(Calendar.DAY_OF_MONTH);
        int mes = calendario.get(Calendar.MONTH) + 1;
        int año = calendario.get(Calendar.YEAR);


        return new Fecha(dia, mes, año);
    }

    public static void main(String[] args) {
        try {
            Fecha fechaValida = new Fecha(29, 2, 2024); 
            System.out.println("Fecha válida: " + fechaValida);

            Fecha fechaInvalida = new Fecha(31, 13, 2022);  
            System.out.println("Fecha inválida: " + fechaInvalida);
        } catch (ExcepcionFechaNoValida e) {
            System.out.println("Error: " + e.getMessage());
        }
    }

}
