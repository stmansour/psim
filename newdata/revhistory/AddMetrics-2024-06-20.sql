/*
There were a total of 73 "GCAM" metrics in the MISubclasses table. Only 10
of these metrics had "_ECON" counterparts. They all needed "_ECON" counterparts.
The query below selected all the "GCAM"metrics that needed "_ECON" counterparts
to be added to the table.

    SELECT t1.Metric
    FROM MISubclasses t1
    WHERE t1.Metric LIKE 'GCAM_%'
        AND t1.Metric NOT LIKE '%_ECON'
        AND NOT EXISTS(
            SELECT 1
            FROM MISubclasses t2
            WHERE t2.Metric = CONCAT(t1.Metric, '_ECON')
        );

The solution set was then formed into the values for the INSERT statement below.
*/

INSERT INTO MISubclasses (
    Name, Metric, Subclass, LocaleType, MetricType, Predictor,
    MinDelta1, MaxDelta1, MinDelta2, MaxDelta2, FitnessW1,
    FitnessW2, HoldWindowPos, HoldWindowNeg)
VALUES
    ('GDELT Name', 'GCAM_C15_129_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C15_130_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C15_132_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C15_133_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C15_205_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C16_21_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C16_37_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C17_13_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C17_31_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C17_34_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C17_38_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_10_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_11_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_2_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_3_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_4_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_5_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_6_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_7_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_8_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C25_9_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C3_3_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C3_4_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C41_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C42_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_10_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_14_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_16_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_17_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_19_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_20_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_21_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_24_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_25_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_26_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C4_6_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C5_32_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C5_33_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C5_34_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C5_35_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C7_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_C7_2_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V10_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V10_2_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V11_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V19_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V19_2_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V19_3_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V21_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V26_1_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_10_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_11_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_2_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_3_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_4_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_5_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_6_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_7_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_8_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_V42_9_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001),
    ('GDELT Name', 'GCAM_WC_ECON', 'LSMInfluencer', 1,0,2, -360,-62,-61,-1,0.6,0.4,0.001,-0.001);